package rules

import (
	"go/ast"
	"strings"

	"go/types"

	"golang.org/x/tools/go/analysis"
)

// ExitOnceRule enforces that `main()` calls os.Exit or log.Fatal* at most once.
type ExitOnceRule struct{}

func (r *ExitOnceRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "exit_once",
		Doc: `prefer a single process-exit call inside main().

When multiple calls to os.Exit or log.Fatal* appear in main(), prefer
delegating program logic to a ` + "run()" + ` function that returns an error
and centralize the exit/print logic in ` + "main()" + ` so there is exactly one
place that terminates the process. This makes cleanup and testing easier.
`,
		Run: r.run,
	}
}

func (r *ExitOnceRule) run(pass *analysis.Pass) (any, error) {
	// Only applies to package main
	if pass.Pkg == nil || pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Find the main function
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Name == nil || fd.Recv != nil {
				continue
			}
			if fd.Name.Name != "main" {
				continue
			}

			// Count exit-like calls in this main function
			count := 0
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// selector expressions like os.Exit or log.Fatal*
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if obj, ok := pass.TypesInfo.Uses[sel.Sel]; ok {
						if fn, ok := obj.(*types.Func); ok && fn.Pkg() != nil {
							pkgPath := fn.Pkg().Path()
							name := fn.Name()
							if (pkgPath == "os" && name == "Exit") || (pkgPath == "log" && strings.HasPrefix(name, "Fatal")) {
								count++
							}
						}
					}
				}

				return true
			})

			if count > 1 {
				pass.Report(analysis.Diagnostic{
					Pos:     fd.Pos(),
					Message: "prefer a single exit/log.Fatal call in main(); delegate logic to a helper that returns an error and centralize process exit in main",
				})
			}
		}
	}

	return nil, nil
}
