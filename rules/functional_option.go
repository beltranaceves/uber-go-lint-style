package rules

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
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
		Run: r.run,
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

			// Count parameters (each field may have multiple names)
			count := 0
			if fd.Type.Params != nil {
				for _, f := range fd.Type.Params.List {
					if len(f.Names) == 0 {
						count++
					} else {
						count += len(f.Names)
					}
				}
			}

			if count >= 3 {
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
