package rules

import (
	"github.com/mgechev/revive/lint"
)

// ErrorTypeRule lints error-type according to Uber Go Style Guide.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#error-type
type ErrorTypeRule struct{}

// Name returns the rule name
func (r *ErrorTypeRule) Name() string {
	return "error-type"
}

// Apply runs the rule against the provided file
func (r *ErrorTypeRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	// TODO: Implement rule logic
	// - Read file.AST to traverse the Go code
	// - Find patterns that violate the style guide
	// - Return lint.Failure for each violation
	return failures
}
