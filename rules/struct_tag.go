package rules

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// StructTagRule reports exported struct fields that lack tags when the struct
// is used with common marshaling functions (e.g. encoding/json.Marshal).
type StructTagRule struct{}

func (r *StructTagRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "struct_tag",
		Doc: `require struct field tags for marshaled structs.

This rule detects exported struct fields without tags when the struct is
passed to common marshaling functions (for example, encoding/json.Marshal).
Use this to ensure the serialized form is explicit and stable across
refactors. The rule uses type information to find marshaling calls.
`,
		Run: r.run,
	}
}

func (r *StructTagRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			obj := pass.TypesInfo.Uses[sel.Sel]
			fn, ok := obj.(*types.Func)
			if !ok || fn.Pkg() == nil {
				return true
			}

			pkgPath := fn.Pkg().Path()
			name := fn.Name()

			// Common marshal functions
			if !(pkgPath == "encoding/json" && (name == "Marshal" || name == "MarshalIndent")) &&
				!(pkgPath == "gopkg.in/yaml.v3" && name == "Marshal") {
				return true
			}

			for _, arg := range call.Args {
				t := pass.TypesInfo.TypeOf(arg)
				if t == nil {
					continue
				}

				for {
					if ptr, ok := t.(*types.Pointer); ok {
						t = ptr.Elem()
						continue
					}
					if named, ok := t.(*types.Named); ok {
						t = named.Underlying()
						continue
					}
					break
				}

				st, ok := t.(*types.Struct)
				if !ok {
					continue
				}

				for i := 0; i < st.NumFields(); i++ {
					f := st.Field(i)
					if !f.Exported() {
						continue
					}
					tag := st.Tag(i)
					if tag == "" {
						pass.Report(analysis.Diagnostic{
							Pos:     f.Pos(),
							Message: fmt.Sprintf("exported field '%s' should have a tag for marshaling", f.Name()),
						})
					}
				}
			}

			return true
		})
	}
	return nil, nil
}
