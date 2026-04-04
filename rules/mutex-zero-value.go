package rules

import (
	"github.com/mgechev/revive/lint"
)

// MutexZeroValueRule lints mutex-zero-value according to Uber Go Style Guide.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#mutex-zero-value
type MutexZeroValueRule struct{}

// Name returns the rule name
func (r *MutexZeroValueRule) Name() string {
	return "mutex-zero-value"
}

// Apply runs the rule against the provided file
func (r *MutexZeroValueRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	// TODO: Implement rule logic
	// - Read file.AST to traverse the Go code
	// - Find patterns that violate the style guide
	// - Return lint.Failure for each violation
	return failures
}
