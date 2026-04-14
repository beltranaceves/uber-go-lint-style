package rules

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	buildssa "golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

// GoroutineInitRule reports any goroutine started inside an init() function,
// including goroutines started indirectly by callees of init() when reachable
// via static callees in SSA.
type GoroutineInitRule struct{}

func (r *GoroutineInitRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "goroutine_init",
		Doc: `do not spawn goroutines in init().

init() functions should not start background goroutines. If a package needs
background work, expose a worker object that manages the goroutine's lifetime
and provides a Shutdown/Close/Stop method for callers to stop it.
This analyzer uses SSA to detect goroutines started indirectly via functions
called from init().`,
		Run:      r.run,
		Requires: []*analysis.Analyzer{buildssa.Analyzer},
	}
}

func (r *GoroutineInitRule) run(pass *analysis.Pass) (any, error) {
	ssaRes := pass.ResultOf[buildssa.Analyzer]
	ssab, _ := ssaRes.(*buildssa.SSA)

	const maxDepth = 10

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Name == nil || fd.Name.Name != "init" || fd.Body == nil {
				return true
			}

			// 1) Direct go statements in the init body
			directGo := false
			ast.Inspect(fd.Body, func(n2 ast.Node) bool {
				if gs, ok := n2.(*ast.GoStmt); ok {
					directGo = true
					pass.Report(analysis.Diagnostic{
						Pos:     gs.Pos(),
						Message: "do not start goroutines in init",
					})
				}
				return true
			})

			// 2) SSA-based search for indirect goroutine starts reachable from this init
			// Skip SSA search if a direct go was already found to avoid duplicate diagnostics.
			if directGo || ssab == nil {
				return true
			}

			// find SSA function for this init decl
			var entryFn *ssa.Function
			if obj := pass.TypesInfo.Defs[fd.Name]; obj != nil {
				for _, f := range ssab.SrcFuncs {
					if f.Object() == obj {
						entryFn = f
						break
					}
				}
			}
			if entryFn == nil {
				return true
			}

			type item struct {
				fn      *ssa.Function
				d       int
				rootPos token.Pos
			}

			q := []item{{fn: entryFn, d: 0, rootPos: 0}}
			visited := make(map[*ssa.Function]bool)
			visited[entryFn] = true

			foundIndirect := false
			var reportPos token.Pos

			for len(q) > 0 && !foundIndirect {
				it := q[0]
				q = q[1:]
				cur := it.fn
				if cur == nil {
					continue
				}

				// scan instructions for Go instructions
				for _, b := range cur.Blocks {
					for _, instr := range b.Instrs {
						if _, ok := instr.(*ssa.Go); ok {
							// report at the rootPos if available, otherwise at the init func
							if it.rootPos != 0 {
								reportPos = it.rootPos
							} else {
								reportPos = fd.Pos()
							}
							foundIndirect = true
							break
						}
					}
					if foundIndirect {
						break
					}
				}
				if foundIndirect {
					break
				}

				if it.d >= maxDepth {
					continue
				}

				// discover static callees and carry the rootPos (first call's pos)
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
						if next := common.StaticCallee(); next != nil && !visited[next] {
							visited[next] = true
							// rootPos: if current has a rootPos, carry it; otherwise use this instr.Pos()
							rp := it.rootPos
							if rp == 0 {
								rp = instr.Pos()
							}
							q = append(q, item{fn: next, d: it.d + 1, rootPos: rp})
						}
					}
				}
			}

			if foundIndirect && reportPos != 0 {
				pass.Report(analysis.Diagnostic{
					Pos:     reportPos,
					Message: "do not start goroutines in init",
				})
			}

			return true
		})
	}
	return nil, nil
}
