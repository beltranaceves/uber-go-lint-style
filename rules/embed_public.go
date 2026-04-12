package rules

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// EmbedPublicRule flags exported structs that embed exported types.
type EmbedPublicRule struct{}

// BuildAnalyzer returns the analyzer for the embed_public rule
func (r *EmbedPublicRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "embed_public",
		Doc: `avoid embedding exported types in exported (public) structs.

This rule detects when a public struct embeds a public type (struct or interface).
Embedding a public type in a public struct leaks implementation details and
limits future type evolution. Prefer keeping the embedded implementation in a
private field and writing explicit delegate methods instead.`,
		Run: r.run,
	}
}

func typeIdentName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return typeIdentName(t.X)
	case *ast.SelectorExpr:
		// pkg.Type -> return Type
		if t.Sel != nil {
			return t.Sel.Name
		}
	}
	return ""
}

func (r *EmbedPublicRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				// Only care about exported (public) types
				if !ast.IsExported(typeSpec.Name.Name) {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok || structType.Fields == nil {
					continue
				}

				for _, field := range structType.Fields.List {
					// Embedded field has no explicit name
					if len(field.Names) != 0 {
						continue
					}

					typeName := typeIdentName(field.Type)
					if typeName == "" {
						continue
					}

					// If the embedded type is exported, report
					if ast.IsExported(typeName) && strings.TrimSpace(typeName) != "" {
						pass.Report(analysis.Diagnostic{
							Pos:     field.Pos(),
							Message: "avoid embedding exported type '" + typeName + "' in exported struct '" + typeSpec.Name.Name + "'",
						})
					}
				}
			}
		}
	}

	return nil, nil
}
