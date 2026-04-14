package rules

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// InitRule flags use of init() functions.
type InitRule struct{}

// BuildAnalyzer returns the analyzer for the init rule
func (r *InitRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "init",
		Doc:  "avoid init() functions; prefer explicit initialization in main or helper functions",
		Run:  r.run,
	}
}

func (r *InitRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if fn.Recv == nil && fn.Name != nil && fn.Name.Name == "init" {
					pass.Report(analysis.Diagnostic{
						Pos:      fn.Pos(),
						End:      fn.End(),
						Category: "init",
						Message:  "avoid init functions; prefer explicit initialization in main or helper functions",
					})
				}
			}
		}
	}

	return nil, nil
}
