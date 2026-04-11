package rules

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// ContainerCapacityRule suggests preallocating capacity for maps and slices
// when they are populated in loops to avoid repeated reallocations.
type ContainerCapacityRule struct{}

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
	// Collect variables created with make(...) that omit a capacity
	// mapVars maps variable name -> call expression node
	mapVars := map[string]*ast.CallExpr{}
	sliceVars := map[string]*ast.CallExpr{}

	for _, file := range pass.Files {
		// Find declarations and assignments with make(...) on RHS
		ast.Inspect(file, func(n ast.Node) bool {
			switch stmt := n.(type) {
			case *ast.AssignStmt:
				for i, rhs := range stmt.Rhs {
					call, ok := rhs.(*ast.CallExpr)
					if !ok {
						continue
					}
					ident, ok := call.Fun.(*ast.Ident)
					if !ok || ident.Name != "make" {
						continue
					}

					// LHS must have an ident to map variable name
					if i >= len(stmt.Lhs) {
						continue
					}
					if lhsIdent, ok := stmt.Lhs[i].(*ast.Ident); ok {
						// map type
						if _, ok := call.Args[0].(*ast.MapType); ok {
							if len(call.Args) < 2 {
								mapVars[lhsIdent.Name] = call
							}
						}
						// slice type
						if _, ok := call.Args[0].(*ast.ArrayType); ok {
							// make([]T, len) -> len(call.Args) == 2 means capacity omitted
							if len(call.Args) < 3 {
								sliceVars[lhsIdent.Name] = call
							}
						}
					}
				}
			case *ast.DeclStmt:
				// var declarations like: var m = make(...)
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
								name := vs.Names[i].Name
								if _, ok := call.Args[0].(*ast.MapType); ok {
									if len(call.Args) < 2 {
										mapVars[name] = call
									}
								}
								if _, ok := call.Args[0].(*ast.ArrayType); ok {
									if len(call.Args) < 3 {
										sliceVars[name] = call
									}
								}
							}
						}
					}
				}
			}
			return true
		})

		if len(mapVars) == 0 && len(sliceVars) == 0 {
			continue
		}

		// Look for loops that populate the collected containers
		reportedMap := map[string]bool{}
		reportedSlice := map[string]bool{}
		ast.Inspect(file, func(n ast.Node) bool {
			switch loop := n.(type) {
			case *ast.RangeStmt:
				// inspect body for assignments to maps
				ast.Inspect(loop.Body, func(n2 ast.Node) bool {
					// m[key] = val
					if asg, ok := n2.(*ast.AssignStmt); ok {
						for _, lhs := range asg.Lhs {
							if idx, ok := lhs.(*ast.IndexExpr); ok {
								if id, ok := idx.X.(*ast.Ident); ok {
									if _, exists := mapVars[id.Name]; exists {
										if !reportedMap[id.Name] {
											pass.Report(analysis.Diagnostic{
												Pos:     mapVars[id.Name].Pos(),
												Message: "preallocate map capacity when populating in a loop",
											})
											reportedMap[id.Name] = true
										}
									}
								}
							}
						}
					}
					// append usage inside range
					if call, ok := n2.(*ast.CallExpr); ok {
						if funIdent, ok := call.Fun.(*ast.Ident); ok && funIdent.Name == "append" {
							if len(call.Args) > 0 {
								if id, ok := call.Args[0].(*ast.Ident); ok {
									if _, exists := sliceVars[id.Name]; exists {
										if !reportedSlice[id.Name] {
											pass.Report(analysis.Diagnostic{
												Pos:     sliceVars[id.Name].Pos(),
												Message: "preallocate slice capacity when appending in a loop",
											})
											reportedSlice[id.Name] = true
										}
									}
								}
							}
						}
					}
					return true
				})
			case *ast.ForStmt:
				// also inspect generic for-loop bodies for append/assign
				ast.Inspect(loop.Body, func(n2 ast.Node) bool {
					if asg, ok := n2.(*ast.AssignStmt); ok {
						for _, lhs := range asg.Lhs {
							if idx, ok := lhs.(*ast.IndexExpr); ok {
								if id, ok := idx.X.(*ast.Ident); ok {
									if _, exists := mapVars[id.Name]; exists {
										if !reportedMap[id.Name] {
											pass.Report(analysis.Diagnostic{
												Pos:     mapVars[id.Name].Pos(),
												Message: "preallocate map capacity when populating in a loop",
											})
											reportedMap[id.Name] = true
										}
									}
								}
							}
						}
					}
					if call, ok := n2.(*ast.CallExpr); ok {
						if funIdent, ok := call.Fun.(*ast.Ident); ok && funIdent.Name == "append" {
							if len(call.Args) > 0 {
								if id, ok := call.Args[0].(*ast.Ident); ok {
									if _, exists := sliceVars[id.Name]; exists {
										if !reportedSlice[id.Name] {
											pass.Report(analysis.Diagnostic{
												Pos:     sliceVars[id.Name].Pos(),
												Message: "preallocate slice capacity when appending in a loop",
											})
											reportedSlice[id.Name] = true
										}
									}
								}
							}
						}
					}
					return true
				})
			}
			return true
		})
	}

	return nil, nil
}
