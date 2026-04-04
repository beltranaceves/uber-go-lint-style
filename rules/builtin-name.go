package rules

import (
	"github.com/mgechev/revive/lint"
)

// BuiltinNameRule lints builtin-name according to Uber Go Style Guide.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#builtin-name
type BuiltinNameRule struct{}

// Name returns the rule name
func (r *BuiltinNameRule) Name() string {
	return "builtin-name"
}

// Apply runs the rule against the provided file
func (r *BuiltinNameRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	// TODO: Implement rule logic
	// - Read file.AST to traverse the Go code
	// - Find patterns that violate the style guide
	// - Return lint.Failure for each violation
	return failures
}
