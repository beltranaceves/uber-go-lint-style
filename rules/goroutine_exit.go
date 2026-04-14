package rules

import (
	"go/ast"
	"go/token"

	"go/types"

	"golang.org/x/tools/go/analysis"
	buildssa "golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

// GoroutineExitRule ensures goroutines started by the system (in main(), init(),
// or TestMain) have a way to be waited on (e.g., a WaitGroup, done channel, or
// an explicit receive/close).
type GoroutineExitRule struct{}

func (r *GoroutineExitRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "goroutine_exit",
		Doc: `wait for goroutines started by the system to exit.

When a goroutine is spawned from a system-managed entrypoint such as
` + "`main()`" + `, ` + "`init()`" + ` or ` + "`TestMain`" + `, the program should
provide a way to wait for that goroutine to finish (for example a
` + "`sync.WaitGroup`" + `, a done channel that is closed, or an explicit
receive from a channel). This analyzer reports ` + "`go`" + ` statements
found in those entrypoints that do not appear to be waited on within the
same function body.
`,
		Run:      r.run,
		Requires: []*analysis.Analyzer{buildssa.Analyzer},
	}
}

func (r *GoroutineExitRule) run(pass *analysis.Pass) (any, error) {
	// Obtain SSA result from buildssa pass
	ssaRes := pass.ResultOf[buildssa.Analyzer]
	ssab, _ := ssaRes.(*buildssa.SSA)
	// ssab (buildssa.SSA) provides SSA info via ssab.SrcFuncs and ssab.Pkg
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Body == nil || fd.Name == nil {
				return true
			}

			// Restrict checks to entrypoints: main, init, TestMain.
			name := fd.Name.Name
			if !(name == "main" || name == "init" || name == "TestMain") {
				return true
			}

			var gos []*ast.GoStmt

			// Track signals in outer function
			outerHasWait := false
			outerHasClose := false
			outerHasReceive := false

			// For each goroutine literal, track whether it signals completion (Done/close/send)
			closureSignals := map[*ast.FuncLit]struct{}{}

			ast.Inspect(fd.Body, func(n2 ast.Node) bool {
				switch t := n2.(type) {
				case *ast.GoStmt:
					gos = append(gos, t)
					// If the go launches a function literal, analyze its body for Done/close/send
					call := t.Call
					if fl, ok := call.Fun.(*ast.FuncLit); ok {
						// Inspect closure for Done, close, and send operations
						ast.Inspect(fl.Body, func(n3 ast.Node) bool {
							switch u := n3.(type) {
							case *ast.CallExpr:
								if sel, ok := u.Fun.(*ast.SelectorExpr); ok {
									if sel.Sel != nil && sel.Sel.Name == "Done" {
										if obj, ok := pass.TypesInfo.Uses[sel.Sel]; ok {
											if fn, ok := obj.(*types.Func); ok && fn.Pkg() != nil {
												if sig, ok := fn.Type().(*types.Signature); ok && sig.Recv() != nil {
													recv := sig.Recv().Type()
													if ptr, ok := recv.(*types.Pointer); ok {
														recv = ptr.Elem()
													}
													if named, ok := recv.(*types.Named); ok {
														if named.Obj() != nil && named.Obj().Pkg() != nil {
															if named.Obj().Pkg().Path() == "sync" && named.Obj().Name() == "WaitGroup" {
																closureSignals[fl] = struct{}{}
															}
														}
													}
												}
											}
										}
									}
								}
								if ident, ok := u.Fun.(*ast.Ident); ok {
									if ident.Name == "close" {
										closureSignals[fl] = struct{}{}
									}
								}
							case *ast.SendStmt:
								closureSignals[fl] = struct{}{}
							}
							return true
						})
					}
				case *ast.CallExpr:
					if sel, ok := t.Fun.(*ast.SelectorExpr); ok {
						if sel.Sel != nil && sel.Sel.Name == "Wait" {
							if obj, ok := pass.TypesInfo.Uses[sel.Sel]; ok {
								if fn, ok := obj.(*types.Func); ok && fn.Pkg() != nil {
									if sig, ok := fn.Type().(*types.Signature); ok && sig.Recv() != nil {
										recv := sig.Recv().Type()
										if ptr, ok := recv.(*types.Pointer); ok {
											recv = ptr.Elem()
										}
										if named, ok := recv.(*types.Named); ok {
											if named.Obj() != nil && named.Obj().Pkg() != nil {
												if named.Obj().Pkg().Path() == "sync" && named.Obj().Name() == "WaitGroup" {
													outerHasWait = true
												}
											}
										}
									}
								}
							}
						}
					}
					if ident, ok := t.Fun.(*ast.Ident); ok {
						if ident.Name == "close" {
							outerHasClose = true
						}
					}
				case *ast.UnaryExpr:
					if t.Op == token.ARROW {
						outerHasReceive = true
					}
				}
				return true
			})

			// Determine if goroutines are properly waited on.
			// Consider waited if outer function has Wait/close/receive, or if the
			// goroutine's function reaches (via callgraph) a Wait function.
			waited := false
			if outerHasWait || outerHasClose || outerHasReceive {
				waited = true
			}

			// (Callgraph-based interprocedural analysis handled below using SSA)

			// If we have SSA available, do a depth-limited interprocedural
			// search following static callees found in SSA instructions.
			if ssab != nil {
				// find the SSA function corresponding to this AST func decl
				var entryFn *ssa.Function
				if obj := pass.TypesInfo.Defs[fd.Name]; obj != nil {
					for _, f := range ssab.SrcFuncs {
						if f.Object() == obj {
							entryFn = f
							break
						}
					}
				}

				// collect all *ssa.Go instructions in the entry function
				var ssaGos []*ssa.Go
				if entryFn != nil {
					for _, b := range entryFn.Blocks {
						for _, instr := range b.Instrs {
							if gop, ok := instr.(*ssa.Go); ok {
								ssaGos = append(ssaGos, gop)
							}
						}
					}
				}

				const maxDepth = 10
				for _, g := range ssaGos {
					if g == nil {
						continue
					}
					callee := g.Call.StaticCallee()
					if callee == nil {
						continue
					}

					// BFS over functions discovered on the fly via static callees
					type item struct {
						fn *ssa.Function
						d  int
					}
					q := []item{{callee, 0}}
					visitedFn := make(map[*ssa.Function]bool)
					visitedFn[callee] = true
					found := false
					for len(q) > 0 && !found {
						it := q[0]
						q = q[1:]
						cur := it.fn
						if cur == nil {
							continue
						}

						// Is this function a Wait receiver of sync.WaitGroup?
						if cur.Name() == "Wait" && cur.Signature != nil && cur.Signature.Recv() != nil {
							recv := cur.Signature.Recv().Type()
							if ptr, ok := recv.(*types.Pointer); ok {
								recv = ptr.Elem()
							}
							if named, ok := recv.(*types.Named); ok {
								if named.Obj() != nil && named.Obj().Pkg() != nil {
									if named.Obj().Pkg().Path() == "sync" && named.Obj().Name() == "WaitGroup" {
										found = true
										break
									}
								}
							}
						}

						if it.d >= maxDepth {
							continue
						}

						// discover static callees from cur by scanning its instructions
						for _, b := range cur.Blocks {
							for _, instr := range b.Instrs {
								var common *ssa.CallCommon
								switch v := instr.(type) {
								case *ssa.Call:
									common = v.Common()
								case *ssa.Defer:
									common = v.Common()
								case *ssa.Go:
									common = v.Common()
								default:
									continue
								}
								if next := common.StaticCallee(); next != nil && !visitedFn[next] {
									visitedFn[next] = true
									q = append(q, item{next, it.d + 1})
								}
							}
						}
					}
					if found {
						waited = true
						break
					}
				}
			}

			if len(gos) > 0 && !waited {
				for _, g := range gos {
					pass.Report(analysis.Diagnostic{
						Pos:     g.Pos(),
						Message: "goroutine started in main/init/TestMain must have a way to wait for it to exit",
					})
				}
			}

			return true
		})
	}

	return nil, nil
}
