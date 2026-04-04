package rules

import (
	"github.com/mgechev/revive/lint"
)

// StructFieldKeyRule lints struct-field-key according to Uber Go Style Guide.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#struct-field-key
type StructFieldKeyRule struct{}

// Name returns the rule name
func (r *StructFieldKeyRule) Name() string {
	return "struct-field-key"
}

// Apply runs the rule against the provided file
func (r *StructFieldKeyRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	// TODO: Implement rule logic
	// - Read file.AST to traverse the Go code
	// - Find patterns that violate the style guide
	// - Return lint.Failure for each violation
	return failures
}
