package rules

import (
	"github.com/mgechev/revive/lint"
)

// LintRule lints lint according to Uber Go Style Guide.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#lint
type LintRule struct{}

// Name returns the rule name
func (r *LintRule) Name() string {
	return "lint"
}

// Apply runs the rule against the provided file
func (r *LintRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	// TODO: Implement rule logic
	// - Read file.AST to traverse the Go code
	// - Find patterns that violate the style guide
	// - Return lint.Failure for each violation
	return failures
}
