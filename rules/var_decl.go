package rules

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// VarDeclRule enforces using short variable declarations for local vars with initializers.
type VarDeclRule struct{}

// BuildAnalyzer returns the analyzer for the var_decl rule
func (r *VarDeclRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "var_decl",
		Doc:  "prefer short variable declarations (:=) for local variables when an initializer is present",
		Run:  r.run,
	}
}

func (r *VarDeclRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// local var declarations appear as DeclStmt containing a GenDecl
			ds, ok := n.(*ast.DeclStmt)
			if !ok {
				return true
			}

			gd, ok := ds.Decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				return true
			}

			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				// Skip declarations without an initializer (e.g., `var s []int`)
				if len(vs.Values) == 0 {
					continue
				}

				// Skip if an explicit type is provided (`var x T = ...`)
				if vs.Type != nil {
					continue
				}

				// Report: prefer ':=' for local var with initializer
				pass.Report(analysis.Diagnostic{
					Pos:      vs.Pos(),
					End:      0,
					Category: "var_decl",
					Message:  "prefer ':=' for local variable declarations with initializers",
				})
			}

			return true
		})
	}

	return nil, nil
}
