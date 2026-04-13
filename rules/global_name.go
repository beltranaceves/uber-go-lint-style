package rules

import (
	"go/ast"
	"go/token"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

// GlobalNameRule enforces that unexported package-level vars and consts are
// prefixed with an underscore to make their global scope obvious.
type GlobalNameRule struct{}

func (r *GlobalNameRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "global_name",
		Doc: `prefix unexported package-level vars and consts with '_'.

This rule flags unexported top-level var and const identifiers that
do not start with '_' to make it clear when they are globals. Exception: unexported
error sentinel variables may use the prefix 'err' without the underscore.
`,
		Run: r.run,
	}
}

func (r *GlobalNameRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || (gen.Tok != token.VAR && gen.Tok != token.CONST) {
				continue
			}

			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for i, name := range vs.Names {
					if name == nil {
						continue
					}
					// ignore blank identifier
					if name.Name == "_" {
						continue
					}

					// exported names are allowed
					first := []rune(name.Name)[0]
					if unicode.IsUpper(first) {
						continue
					}

					// already prefixed with underscore is acceptable
					if len(name.Name) > 0 && name.Name[0] == '_' {
						continue
					}

					// exception: allow unexported error sentinel vars that start with "err"
					if gen.Tok == token.VAR && pass.TypesInfo != nil && i < len(vs.Values) {
						t := pass.TypesInfo.TypeOf(vs.Values[i])
						if isErrorInterface(t) {
							if len(name.Name) >= 3 && name.Name[:3] == "err" {
								continue
							}
						}
					}

					// report
					pass.Report(analysis.Diagnostic{
						Pos:     name.Pos(),
						Message: "unexported package-level identifier '" + name.Name + "' should be prefixed with '_'",
					})
				}
			}
		}
	}
	return nil, nil
}
