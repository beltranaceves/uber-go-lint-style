package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// ErrorTypeRule enforces correct error declaration choices.
type ErrorTypeRule struct{}

func (r *ErrorTypeRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "error_type",
		Doc: `ensure exported package-level error variables are declared appropriately.

This rule warns when an exported package-level var is initialized with fmt.Errorf (a dynamic error string).
Exported errors should be a top-level static error created with errors.New, or a custom error type when callers need to match the error via errors.Is/errors.As.`,
		Run: r.run,
	}
}

func (r *ErrorTypeRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range vs.Names {
					if !ast.IsExported(name.Name) {
						continue
					}
					if len(vs.Values) <= i {
						continue
					}
					val := vs.Values[i]
					call, ok := val.(*ast.CallExpr)
					if !ok {
						continue
					}
					// Detect calls to fmt.Errorf using type information.
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if obj, ok := pass.TypesInfo.Uses[sel.Sel]; ok {
							if fn, ok := obj.(*types.Func); ok {
								if fn.Pkg() != nil && fn.Pkg().Path() == "fmt" && fn.Name() == "Errorf" {
									pass.Report(analysis.Diagnostic{
										Pos:     name.Pos(),
										Message: "exported error variable is created by fmt.Errorf; export a top-level static error (errors.New) or use a custom error type instead",
									})
								}
							}
						}
					}
				}
			}
		}
	}
	return nil, nil
}
