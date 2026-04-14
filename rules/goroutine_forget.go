package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// GoroutineForgetRule detects fire-and-forget goroutines that contain
// infinite loops and therefore lack a way to stop and wait for them to exit.
type GoroutineForgetRule struct{}

func (r *GoroutineForgetRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "goroutine_forget",
		Doc: `detects goroutines that are started and never given a way to stop.

This rule flags ` + "`go`" + ` statements that launch anonymous function
literals or local functions that contain an infinite loop (for { ... } or
` + "`for true {}`" + `). Such goroutines are effectively fire-and-forget: they
will run until the process exits unless there is explicit stop signaling
and a way for callers to wait for termination. Prefer using a stop channel,
` + "`context.Context`" + ` with cancellation, or a ` + "`sync.WaitGroup`" + ` and expose a
way to wait for the goroutine to exit.

This analyzer is intentionally conservative and uses heuristics; it may not
catch every case and may report false positives in complex control flow. To
detect goroutine leaks at runtime in tests, use go.uber.org/goleak.
`,
		Run: r.run,
	}
}

func (r *GoroutineForgetRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			goStmt, ok := n.(*ast.GoStmt)
			if !ok {
				return true
			}

			call := goStmt.Call

			// We analyze either anonymous function literals or simple named
			// function calls in the same file. For named functions we try to
			// locate the declaration in this file and inspect its body.
			reported := false

			// helper to inspect a block for infinite loops
			inspectBody := func(body *ast.BlockStmt) bool {
				if body == nil {
					return false
				}
				found := false
				ast.Inspect(body, func(n2 ast.Node) bool {
					if found {
						return false
					}
					if forStmt, ok := n2.(*ast.ForStmt); ok {
						// Heuristics: treat `for {}` and `for true {}` as infinite
						if (forStmt.Init == nil && forStmt.Cond == nil && forStmt.Post == nil) || isCondTrue(forStmt.Cond) {
							if hasSelectWithStopReturn(forStmt) {
								return true
							}
							found = true
							return false
						}
					}
					return true
				})
				return found
			}

			// Case A: anonymous function literal
			if funLit, ok := call.Fun.(*ast.FuncLit); ok {
				if inspectBody(funLit.Body) {
					reported = true
				}
			}

			// Case B: named function call in same file (go foo())
			if !reported {
				if ident, ok := call.Fun.(*ast.Ident); ok {
					// Find local FuncDecl with this name in the same file
					for _, decl := range file.Decls {
						if fd, ok := decl.(*ast.FuncDecl); ok {
							if fd.Name != nil && fd.Name.Name == ident.Name {
								if inspectBody(fd.Body) {
									reported = true
									break
								}
							}
						}
					}
				}
			}

			if reported {
				pass.Report(analysis.Diagnostic{
					Pos:     goStmt.Pos(),
					Message: "fire-and-forget goroutine contains an infinite loop; provide stop signaling and a way to wait for the goroutine to exit (consider testing for leaks with go.uber.org/goleak)",
				})
			}

			return true
		})
	}
	return nil, nil
}

// hasSelectWithStopReturn checks whether the for-loop body contains a select
// statement with a receive case ("<-ident") whose body contains a return
// statement. This is a heuristic to identify loops that can be signaled to
// stop.
func hasSelectWithStopReturn(forStmt *ast.ForStmt) bool {
	found := false
	ast.Inspect(forStmt.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		sel, ok := n.(*ast.SelectStmt)
		if !ok {
			return true
		}
		for _, s := range sel.Body.List {
			cc, ok := s.(*ast.CommClause)
			if !ok || cc.Comm == nil {
				continue
			}
			// Expect Comm to be an ExprStmt containing a UnaryExpr with Op == ARROW
			exprStmt, ok := cc.Comm.(*ast.ExprStmt)
			if !ok {
				continue
			}
			unary, ok := exprStmt.X.(*ast.UnaryExpr)
			if !ok || unary.Op != token.ARROW {
				continue
			}
			// If the clause body contains a return statement, consider it a stop
			// signal that exits the goroutine.
			hasReturn := false
			ast.Inspect(&ast.BlockStmt{List: cc.Body}, func(n2 ast.Node) bool {
				if hasReturn {
					return false
				}
				if _, ok := n2.(*ast.ReturnStmt); ok {
					hasReturn = true
					return false
				}
				return true
			})
			if hasReturn {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// isCondTrue reports whether an expression is a boolean `true` literal.
func isCondTrue(expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.Ident:
		return strings.EqualFold(e.Name, "true")
	case *ast.BasicLit:
		return strings.EqualFold(strings.TrimSpace(e.Value), "true")
	}
	return false
}
