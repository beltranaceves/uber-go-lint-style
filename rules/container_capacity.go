package rules

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// ContainerCapacityRule suggests preallocating capacity for maps and slices
// when they are populated in loops to avoid repeated reallocations.
type ContainerCapacityRule struct{}

/*
Pseudocode - how this analyzer works

for each file in pass.Files:
	// phase 1: collect make(...) calls that omit capacity
	for each node in file:
		if node is assignment or var decl and rhs is make(...):
			obj := pass.TypesInfo.ObjectOf(lhs)
			t := pass.TypesInfo.TypeOf(call)
			if t is map and no capacity arg: mapVars[obj] = call
			if t is slice and no capacity arg: sliceVars[obj] = call

	if mapVars and sliceVars are both empty: continue

	// phase 2: find loops that populate those containers
	reported := empty set
	for each node in file:
		if node is ForStmt or RangeStmt:
			inspect loop body:
				if stmt is assignment with IndexExpr on LHS:
					target := resolve object of IndexExpr.X via pass.TypesInfo
					if target in mapVars and not reported: report at mapVars[target]; mark reported
				if stmt is CallExpr and function is append:
					target := resolve object of first arg via pass.TypesInfo
					if target in sliceVars and not reported: report at sliceVars[target]; mark reported

Notes:
- Uses pass.TypesInfo to resolve identifiers/selectors to types.Object so matches are
	robust across scopes and named types.
- Reports point at the original make(...) call to guide preallocation.
*/

// BuildAnalyzer returns the analyzer for the container_capacity rule
func (r *ContainerCapacityRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "container_capacity",
		Doc: `preallocate capacity for maps and slices when populating in loops

This rule detects common patterns where a map or slice is created without
an explicit capacity and then populated via a loop (range or append). In
those cases preallocating capacity (for maps, using make(map[K]V, n); for
slices, using make([]T, 0, n)) avoids repeated growth and allocations.`,
		Run: r.run,
	}
}

func (r *ContainerCapacityRule) run(pass *analysis.Pass) (any, error) {
	// Iterate files: collect make(...) calls then inspect loop bodies
	for _, file := range pass.Files {
		// Collect variables created by `make` without capacity
		mapVars, sliceVars := collectMakeVars(pass, file)
		if len(mapVars) == 0 && len(sliceVars) == 0 {
			continue
		}
		// Track reported objects to avoid duplicate diagnostics
		reportedMap := map[types.Object]bool{}
		reportedSlice := map[types.Object]bool{}
		// Walk file AST and handle loop statements
		ast.Inspect(file, func(n ast.Node) bool {
			switch loop := n.(type) {
			case *ast.RangeStmt:
				inspectLoopBody(pass, loop.Body, mapVars, sliceVars, reportedMap, reportedSlice)
			case *ast.ForStmt:
				inspectLoopBody(pass, loop.Body, mapVars, sliceVars, reportedMap, reportedSlice)
			}
			return true
		})
	}

	return nil, nil
}

// collectMakeVars finds `make(...)` calls in the given file that create maps
// or slices without an explicit capacity and returns two maps keyed by the
// declared object: (mapVars, sliceVars).
func collectMakeVars(pass *analysis.Pass, file *ast.File) (map[types.Object]*ast.CallExpr, map[types.Object]*ast.CallExpr) {
	mapVars := map[types.Object]*ast.CallExpr{}
	sliceVars := map[types.Object]*ast.CallExpr{}

	// Walk AST to find make(...) calls used in assignments
	ast.Inspect(file, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			// Handle assignment statements: lhs = make(...)
			for i, rhs := range stmt.Rhs {
				call, ok := rhs.(*ast.CallExpr)
				if !ok {
					continue
				}
				ident, ok := call.Fun.(*ast.Ident)
				if !ok || ident.Name != "make" {
					continue
				}

				if i >= len(stmt.Lhs) {
					continue
				}
				if lhsIdent, ok := stmt.Lhs[i].(*ast.Ident); ok {
					if pass.TypesInfo == nil {
						continue
					}
					obj := pass.TypesInfo.ObjectOf(lhsIdent)
					if obj == nil {
						continue
					}
					t := pass.TypesInfo.TypeOf(call)
					if t == nil {
						continue
					}
					switch ut := t.Underlying().(type) {
					case *types.Map:
						if len(call.Args) < 2 {
							mapVars[obj] = call
						}
					case *types.Slice:
						if len(call.Args) < 3 {
							sliceVars[obj] = call
						}
					default:
						_ = ut
					}
				}
			}
		// Handle var declarations like: var x = make(...)
		case *ast.DeclStmt:
			if gd, ok := stmt.Decl.(*ast.GenDecl); ok {
				for _, spec := range gd.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for i, value := range vs.Values {
							call, ok := value.(*ast.CallExpr)
							if !ok {
								continue
							}
							ident, ok := call.Fun.(*ast.Ident)
							if !ok || ident.Name != "make" {
								continue
							}
							if i >= len(vs.Names) {
								continue
							}
							name := vs.Names[i]
							if pass.TypesInfo == nil {
								continue
							}
							obj := pass.TypesInfo.ObjectOf(name)
							if obj == nil {
								continue
							}
							t := pass.TypesInfo.TypeOf(call)
							if t == nil {
								continue
							}
							switch ut := t.Underlying().(type) {
							case *types.Map:
								if len(call.Args) < 2 {
									mapVars[obj] = call
								}
							case *types.Slice:
								if len(call.Args) < 3 {
									sliceVars[obj] = call
								}
							default:
								_ = ut
							}
						}
					}
				}
			}
		}
		return true
	})

	return mapVars, sliceVars
}

// objFromExpr resolves an identifier or selector expression to its
// corresponding types.Object using pass.TypesInfo when possible.
func objFromExpr(pass *analysis.Pass, e ast.Expr) types.Object {
	if pass == nil || pass.TypesInfo == nil || e == nil {
		return nil
	}
	switch x := e.(type) {
	case *ast.Ident:
		return pass.TypesInfo.ObjectOf(x)
	case *ast.SelectorExpr:
		if o := pass.TypesInfo.ObjectOf(x.Sel); o != nil {
			return o
		}
		if sel := pass.TypesInfo.Selections[x]; sel != nil {
			return sel.Obj()
		}
	}
	return nil
}

// inspectLoopBody walks a loop body and reports diagnostics when it finds
// map index assignments or append calls that target recorded containers.
func inspectLoopBody(pass *analysis.Pass, body *ast.BlockStmt, mapVars, sliceVars map[types.Object]*ast.CallExpr, reportedMap, reportedSlice map[types.Object]bool) {
	if body == nil {
		return
	}
	ast.Inspect(body, func(n ast.Node) bool {
		if asg, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range asg.Lhs {
				if idx, ok := lhs.(*ast.IndexExpr); ok {
					if obj := objFromExpr(pass, idx.X); obj != nil {
						if call, exists := mapVars[obj]; exists && !reportedMap[obj] {
							pass.Report(analysis.Diagnostic{
								Pos:     call.Pos(),
								Message: "preallocate map capacity when populating in a loop",
							})
							reportedMap[obj] = true
						}
					}
				}
			}
		}

		if call, ok := n.(*ast.CallExpr); ok {
			if funIdent, ok := call.Fun.(*ast.Ident); ok && funIdent.Name == "append" {
				if len(call.Args) > 0 {
					if obj := objFromExpr(pass, call.Args[0]); obj != nil {
						if c, exists := sliceVars[obj]; exists && !reportedSlice[obj] {
							pass.Report(analysis.Diagnostic{
								Pos:     c.Pos(),
								Message: "preallocate slice capacity when appending in a loop",
							})
							reportedSlice[obj] = true
						}
					}
				}
			}
		}

		return true
	})
}
