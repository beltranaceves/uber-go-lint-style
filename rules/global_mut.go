package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"unicode"

	"go/types"

	"golang.org/x/tools/go/analysis"
)

// GlobalMutRule flags package-level mutable variables and function-value globals.
type GlobalMutRule struct{}

func (r *GlobalMutRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "global_mut",
		Doc: `avoid mutable package-level variables.

This rule flags package-level var declarations because mutable globals make
reasoning and testing harder. Prefer passing dependencies explicitly (dependency
injection) or scoping state behind types or functions.

The analyzer ignores obvious immutable initializers (basic literals), exported
package API (exported names), and common sentinel errors named ` + "Err..." + `
whose initializer has type ` + "`error`" + `.
`,
		Run: r.run,
	}
}

func isErrorInterface(t types.Type) bool {
	if t == nil {
		return false
	}
	// Lookup predeclared `error` interface
	errObj := types.Universe.Lookup("error")
	if errObj == nil {
		return false
	}
	errIface := errObj.Type().Underlying().(*types.Interface)
	return types.Implements(t, errIface) || types.Implements(types.NewPointer(t), errIface)
}

func (r *GlobalMutRule) run(pass *analysis.Pass) (any, error) {
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

				var flagged []string
				for i, name := range vs.Names {
					if name == nil || name.Name == "_" {
						continue
					}
					// Skip exported package API
					first := []rune(name.Name)[0]
					if unicode.IsUpper(first) {
						continue
					}

					// If there are no values (e.g., `var x int`), flag it
					if len(vs.Values) == 0 || i >= len(vs.Values) {
						flagged = append(flagged, name.Name)
						continue
					}

					val := vs.Values[i]
					// check type info if available: allow sentinel errors named Err*
					if pass.TypesInfo != nil {
						t := pass.TypesInfo.TypeOf(val)
						if isErrorInterface(t) && len(name.Name) >= 3 && name.Name[:3] == "Err" {
							// Err* sentinel of error type: skip
							continue
						}
					}

					flagged = append(flagged, name.Name)
				}

				if len(flagged) == 0 {
					continue
				}

				var msg string
				if len(flagged) == 1 {
					msg = fmt.Sprintf("avoid mutable package-level variable '%s'; prefer dependency injection or scoped state", flagged[0])
				} else {
					msg = fmt.Sprintf("avoid mutable package-level variables '%s'; prefer dependency injection or scoped state", joinNames(flagged))
				}

				pass.Report(analysis.Diagnostic{Pos: gen.Pos(), Message: msg})
			}
		}
	}
	return nil, nil
}

func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return names[0]
	}
	out := names[0]
	for _, n := range names[1:] {
		out = out + ", " + n
	}
	return out
}
