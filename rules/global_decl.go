package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// GlobalDeclRule enforces top-level variable declaration style.
type GlobalDeclRule struct{}

// BuildAnalyzer returns the analyzer for global-decl rule
func (r *GlobalDeclRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "global_decl",
		Doc: `At the top level, prefer using the 'var' keyword without an explicit type
when the initializer already provides the type. Specify the type only when the
initializer's type differs from the desired declared type.

This rule flags top-level ` + "var" + ` declarations that include an explicit
type and an initializer whose type is identical to the declared type.
`,
		Run: r.run,
	}
}

func (r *GlobalDeclRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}

			for _, spec := range gen.Specs {
				valSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				// Only consider specs that have an explicit type and at least one value.
				if valSpec.Type == nil || len(valSpec.Values) == 0 {
					continue
				}

				declaredType := pass.TypesInfo.TypeOf(valSpec.Type)
				if declaredType == nil {
					continue
				}

				// Choose value index: if values match names one-to-one, use corresponding value,
				// otherwise use the first value (common case: single RHS expression).
				for i, name := range valSpec.Names {
					valueIndex := 0
					if len(valSpec.Values) == len(valSpec.Names) {
						valueIndex = i
					}

					val := valSpec.Values[valueIndex]
					valType := pass.TypesInfo.TypeOf(val)
					if valType == nil {
						continue
					}

					if types.Identical(declaredType, valType) {
						pass.Report(analysis.Diagnostic{
							Pos:     name.Pos(),
							Message: "omit the explicit type in top-level var; use var name = expr instead",
						})
					}
				}
			}
		}
	}

	return nil, nil
}
