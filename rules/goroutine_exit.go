package rules

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
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
		Run: r.run,
	}
}

func (r *GoroutineExitRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Body == nil || fd.Name == nil {
				return true
			}

			name := fd.Name.Name
			if !(name == "main" || name == "init" || name == "TestMain") {
				return true
			}

			var gos []*ast.GoStmt
			waited := false

			ast.Inspect(fd.Body, func(n2 ast.Node) bool {
				switch t := n2.(type) {
				case *ast.GoStmt:
					gos = append(gos, t)
				case *ast.CallExpr:
					// Detect wg.Wait() or close(done)
					if sel, ok := t.Fun.(*ast.SelectorExpr); ok {
						if sel.Sel != nil && sel.Sel.Name == "Wait" {
							waited = true
						}
					}
					if ident, ok := t.Fun.(*ast.Ident); ok {
						if ident.Name == "close" {
							waited = true
						}
					}
				case *ast.UnaryExpr:
					// <-ch or receive expressions
					if t.Op == token.ARROW {
						waited = true
					}
				}
				return true
			})

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
