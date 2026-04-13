package rules

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	buildssa "golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

// FunctionalOptionRule suggests using the functional options pattern for public
// APIs that already have three or more parameters.
type FunctionalOptionRule struct{}

func (r *FunctionalOptionRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "functional_option",
		Doc: `suggest using the functional options pattern for public APIs with many parameters.

This rule flags exported functions or methods that have three or more parameters
and recommends the functional options pattern for optional/expandable arguments.
`,
		Run:      r.run,
		Requires: []*analysis.Analyzer{buildssa.Analyzer},
	}
}

func (r *FunctionalOptionRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Name == nil {
				return true
			}

			// Only consider exported functions/methods
			if !ast.IsExported(fd.Name.Name) {
				return true
			}

			// Build param name list (generate placeholders for unnamed params)
			var paramNames []string
			if fd.Type.Params != nil {
				for i, f := range fd.Type.Params.List {
					if len(f.Names) == 0 {
						// unnamed parameter - create synthetic name
						paramNames = append(paramNames, "__unnamed_param_"+string(rune(i)))
					} else {
						for _, n := range f.Names {
							paramNames = append(paramNames, n.Name)
						}
					}
				}
			}

			total := len(paramNames)
			if total < 3 {
				return true
			}

			// Collect earliest use position for each parameter (token.NoPos means unused)
			_ = token.NoPos

			// Use a combination of AST and SSA/CFG info to determine whether any parameter
			// is optional in the sense that there exists a path to a return block
			// that does not use the parameter. This is more accurate than a
			// purely syntactic heuristic.
			// First, collect syntactic (AST) uses of parameter idents so we can
			// detect simple usages that SSA might optimize away (for example,
			// assignments to the blank identifier).
			usedInAST := make([]bool, total)
			earliestAST := make([]token.Pos, total)
			for i := range earliestAST {
				earliestAST[i] = token.NoPos
			}
			var returnASTs []token.Pos
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.Ident:
					for i, pname := range paramNames {
						if x.Name == pname {
							usedInAST[i] = true
							if earliestAST[i] == token.NoPos || x.Pos() < earliestAST[i] {
								earliestAST[i] = x.Pos()
							}
						}
					}
				case *ast.ReturnStmt:
					returnASTs = append(returnASTs, x.Pos())
				}
				return true
			})

			hasOptional := false
			ssaRes := pass.ResultOf[buildssa.Analyzer]
			ssab, ok := ssaRes.(*buildssa.SSA)
			if !ok || ssab == nil {
				pass.Report(analysis.Diagnostic{
					Pos:     fd.Name.Pos(),
					Message: "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments",
				})
				return true
			}

			// Find the corresponding *ssa.Function for this declaration
			var target *ssa.Function
			if obj := pass.TypesInfo.Defs[fd.Name]; obj != nil {
				for _, f := range ssab.SrcFuncs {
					if f.Object() == obj {
						target = f
						break
					}
				}
			}

			if target == nil {
				// Not all declarations have SSA functions (e.g., methods in other files).
				pass.Report(analysis.Diagnostic{
					Pos:     fd.Name.Pos(),
					Message: "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments",
				})
				return true
			}

			// For each parameter, collect blocks that use it.
			paramUses := make([]map[*ssa.BasicBlock]bool, len(target.Params))
			for i := range paramUses {
				paramUses[i] = make(map[*ssa.BasicBlock]bool)
			}

			usedAny := make([]bool, len(target.Params))
			for i, pv := range target.Params {
				if pv == nil {
					continue
				}
				refs := pv.Referrers()
				if refs == nil || len(*refs) == 0 {
					// If SSA shows no referrers, use AST information. If the AST
					// shows no use, parameter is unused and therefore optional.
					if i < len(usedInAST) && usedInAST[i] {
						// If there is any return that appears before the earliest
						// AST use of this parameter, consider it optional.
						if earliestAST[i] != token.NoPos {
							for _, rpos := range returnASTs {
								if rpos < earliestAST[i] {
									// parameter is optional
									hasOptional = true
									break
								}
							}
							if hasOptional {
								break
							}
						}
						usedAny[i] = true
					} else {
						usedAny[i] = false
					}
					continue
				}
				for _, ref := range *refs {
					if ref == nil || ref.Block() == nil {
						continue
					}
					paramUses[i][ref.Block()] = true
					usedAny[i] = true
				}
			}

			// Identify return blocks
			var returnBlocks []*ssa.BasicBlock
			for _, b := range target.Blocks {
				if len(b.Instrs) == 0 {
					continue
				}
				switch b.Instrs[len(b.Instrs)-1].(type) {
				case *ssa.Return:
					returnBlocks = append(returnBlocks, b)
				}
			}

			// Graph reachability: for each parameter, see if there's a path from
			// entry to any return block that avoids blocks where the parameter is used.
			entry := target.Blocks[0]
			// (no debug prints)

			for i := 0; i < len(paramUses); i++ {
				// If SSA showed no referrers and we conservatively marked it used,
				// but there are no concrete use blocks, assume parameter is used
				// and not optional.
				if usedAny[i] && len(paramUses[i]) == 0 {
					continue
				}

				// If there are no uses at all (and SSA showed no referrers), parameter is optional
				if !usedAny[i] && len(paramUses[i]) == 0 {
					hasOptional = true
					break
				}

				// BFS from entry avoiding use-blocks
				visited := make(map[*ssa.BasicBlock]bool)
				queue := []*ssa.BasicBlock{entry}
				visited[entry] = true
				for len(queue) > 0 && !hasOptional {
					cur := queue[0]
					queue = queue[1:]
					// If current is a return block, we've found a return reachable without using param
					skip := false
					if paramUses[i][cur] {
						skip = true
					}
					if skip {
						continue
					}
					for _, rb := range returnBlocks {
						if rb == cur {
							hasOptional = true
							break
						}
					}
					if hasOptional {
						break
					}
					for _, succ := range cur.Succs {
						if !visited[succ] {
							visited[succ] = true
							queue = append(queue, succ)
						}
					}
				}
				if hasOptional {
					break
				}
			}

			if hasOptional {
				pass.Report(analysis.Diagnostic{
					Pos:     fd.Name.Pos(),
					Message: "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments",
				})
			}

			return true
		})
	}
	return nil, nil
}
