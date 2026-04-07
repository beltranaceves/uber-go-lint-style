package rules

import (
	"go/ast"
	"unicode"

	"github.com/mgechev/revive/lint"
)

// GlobalMutRule discourages mutable global variables.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#global-mut
type GlobalMutRule struct{}

// Name returns the rule name
func (r *GlobalMutRule) Name() string {
	return "global-mut"
}

// Apply runs the rule against the provided file
func (r *GlobalMutRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	if file == nil || file.AST == nil {
		return failures
	}

	// Check top-level declarations (excluding imported packages)
	for _, decl := range file.AST.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok.String() != "var" {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, name := range valueSpec.Names {
				// Flag exported mutable globals
				if isExportedGlobal(name.Name) {
					failures = append(failures, lint.Failure{
						Failure:    "Avoid mutable global variables. Use package functions or methods instead.",
						Node:       name,
						Confidence: 0.9,
					})
				}
			}
		}
	}

	return failures
}

func isExportedGlobal(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Exported if starts with uppercase
	return unicode.IsUpper(rune(name[0]))
}
