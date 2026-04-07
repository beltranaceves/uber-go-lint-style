package rules

import (
	"go/ast"
	"unicode"

	"github.com/mgechev/revive/lint"
)

// StructEmbedRule ensures struct embedding is explicit about exported fields.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#struct-embed
type StructEmbedRule struct{}

// Name returns the rule name
func (r *StructEmbedRule) Name() string {
	return "struct-embed"
}

// Apply runs the rule against the provided file
func (r *StructEmbedRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	if file == nil || file.AST == nil {
		return failures
	}

	ast.Inspect(file.AST, func(n ast.Node) bool {
		// Check struct definitions
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		if structType.Fields == nil {
			return true
		}

		for _, field := range structType.Fields.List {
			// Check for embedded fields (no name)
			if len(field.Names) == 0 {
				// This is embedded - check if it's a basic type or a named type
				if ident, ok := field.Type.(*ast.Ident); ok {
					// Embedded basic types should use explicit syntax
					if isBasicType(ident.Name) {
						failures = append(failures, lint.Failure{
							Failure:    "Embed basic types explicitly with a named field, not as anonymous embedding.",
							Node:       field,
							Confidence: 0.6,
						})
					}
				}
			}
		}

		return true
	})

	return failures
}

func isBasicType(name string) bool {
	switch name {
	case "string", "int", "int32", "int64", "uint", "uint32", "uint64",
		"float32", "float64", "bool", "byte", "rune", "error":
		return true
	}
	return false
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}
