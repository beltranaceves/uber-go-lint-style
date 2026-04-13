package rules

import (
	"go/ast"
	"strings"

	"go/types"

	"golang.org/x/tools/go/analysis"
)

// ExitMainRule enforces calling os.Exit or log.Fatal* only from main().
type ExitMainRule struct{}

func (r *ExitMainRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "exit_main",
		Doc: `call os.Exit or log.Fatal* only in main().

Call one of os.Exit or log.Fatal* only inside the program's main() function
in package main. Other functions should return errors instead of exiting the
process so callers (and tests) can handle failures.`,
		Run: r.run,
	}
}

func (r *ExitMainRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				return true
			}

			ast.Inspect(fd.Body, func(n2 ast.Node) bool {
				// If the call is a plain panic(...) invocation, detect via Ident.
				if call, ok := n2.(*ast.CallExpr); ok {
					// Skip test files
					pos := pass.Fset.Position(call.Pos())
					if strings.HasSuffix(pos.Filename, "_test.go") {
						return true
					}

					// case 1: selector calls like os.Exit or log.Fatal*
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if obj, ok := pass.TypesInfo.Uses[sel.Sel]; ok {
							if fn, ok := obj.(*types.Func); ok && fn.Pkg() != nil {
								pkgPath := fn.Pkg().Path()
								name := fn.Name()
								if (pkgPath == "os" && name == "Exit") || (pkgPath == "log" && strings.HasPrefix(name, "Fatal")) {
									if !(pass.Pkg.Name() == "main" && fd.Name != nil && fd.Name.Name == "main") {
										pass.Report(analysis.Diagnostic{
											Pos:     call.Pos(),
											Message: "call to os.Exit or log.Fatal functions should only be in main(); return an error instead",
										})
									}
								}
							}
						}
						return true
					}

					// case 2: direct panic(...) calls (Ident)
					if ident, ok := call.Fun.(*ast.Ident); ok {
						if ident.Name == "panic" {
							// Consider panic outside main() a violation. Allow in main.
							if !(pass.Pkg.Name() == "main" && fd.Name != nil && fd.Name.Name == "main") {
								pass.Report(analysis.Diagnostic{
									Pos:     call.Pos(),
									Message: "panic should not be used to exit programs; return an error instead",
								})
							}
						}
					}
				}

				return true
			})

			return true
		})
	}

	return nil, nil
}
