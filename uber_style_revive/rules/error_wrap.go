package rules

import (
	"go/ast"
	"strings"

	"github.com/mgechev/revive/lint"
)

// ErrorWrapRule enforces that errors are wrapped with context.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#error-wrap
type ErrorWrapRule struct{}

// Name returns the rule name
func (r *ErrorWrapRule) Name() string {
	return "error-wrap"
}

// Apply runs the rule against the provided file
func (r *ErrorWrapRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	if file == nil || file.AST == nil {
		return failures
	}

	// Check function returns
	ast.Inspect(file.AST, func(n ast.Node) bool {
		// Look for function calls that return errors without wrapping
		callExpr, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		// Check if returning a bare error result from another function
		if len(callExpr.Results) == 1 {
			if callExpr1, ok := callExpr.Results[0].(*ast.CallExpr); ok {
				// Check if calling something that looks like an error-returning function
				if sel, ok := callExpr1.Fun.(*ast.SelectorExpr); ok {
					if sel.Sel.Name == "Errorf" || sel.Sel.Name == "WithMessage" {
						return true
					}
					// Flag bare returns from error sources (like sql.QueryRow.Scan)
					if strings.Contains(sel.Sel.Name, "Scan") || strings.Contains(sel.Sel.Name, "Parse") {
						failures = append(failures, lint.Failure{
							Failure:    "Error not wrapped with context. Use fmt.Errorf() or errors.WithMessage() to wrap errors.",
							Node:       callExpr,
							Confidence: 0.8,
						})
					}
				}
			}
		}

		return true
	})

	return failures
}
