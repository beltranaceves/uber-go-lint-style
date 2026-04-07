package rules

import (
	"go/ast"
	"strings"

	"github.com/mgechev/revive/lint"
)

// ErrorNameRule enforces error naming conventions.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#error-name
type ErrorNameRule struct{}

// Name returns the rule name
func (r *ErrorNameRule) Name() string {
	return "error-name"
}

// Apply runs the rule against the provided file
func (r *ErrorNameRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	if file == nil || file.AST == nil {
		return failures
	}

	ast.Inspect(file.AST, func(n ast.Node) bool {
		// Check var declarations
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		for _, spec := range genDecl.Specs {
			// Check value specs (var declarations)
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Check if initializing an error type
			for i, name := range valueSpec.Names {
				if i < len(valueSpec.Values) {
					// Check if the value is a function call that returns an error
					if callExpr, ok := valueSpec.Values[i].(*ast.CallExpr); ok {
						if isErrorReturningCall(callExpr) {
							// Variables holding errors should be named like "err..." or "...Err"
							if !strings.Contains(name.Name, "err") && !strings.Contains(name.Name, "Err") {
								failures = append(failures, lint.Failure{
									Failure:    "Error variable should be named with 'err' or 'Err' (e.g., 'err', 'parseErr').",
									Node:       name,
									Confidence: 0.8,
								})
							}
						}
					}
				}
			}
		}

		return true
	})

	return failures
}

func isErrorReturningCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		name := sel.Sel.Name
		// Common error-returning functions
		return strings.Contains(name, "Parse") || strings.Contains(name, "Unmarshal") ||
			strings.Contains(name, "ReadFile") || strings.Contains(name, "WriteFile") ||
			strings.Contains(name, "Query") || name == "Scan" || name == "Error"
	}
	return false
}
