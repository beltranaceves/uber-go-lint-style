package rules

import (
	"github.com/mgechev/revive/lint"
)

// ExampleRule is a template for implementing custom revive rules.
// This demonstrates the basic structure that all custom rules should follow.
//
// To implement a custom rule:
// 1. Create a new struct that embeds or implements the lint.Rule interface
// 2. Implement the Name() method to return the rule identifier
// 3. Implement the Apply() method to perform the linting logic
// 4. Register the rule in the revive.toml configuration file
type ExampleRule struct{}

// Name returns the rule name/identifier
func (r *ExampleRule) Name() string {
	return "example-rule"
}

// Apply runs the rule against the provided AST node
func (r *ExampleRule) Apply(file *lint.File, arguments lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	// Example: Check for a simple pattern in the code
	// This is a placeholder - replace with actual linting logic

	// Return all found failures
	return failures
}

// Example of a more complex rule structure with configuration:
//
// type CustomRule struct {
//     enabled bool
//     config  map[string]interface{}
// }
//
// func (r *CustomRule) Name() string {
//     return "custom-rule-name"
// }
//
// func (r *CustomRule) Apply(file *lint.File, arguments lint.Arguments) []lint.Failure {
//     var failures []lint.Failure
//     // Implement rule logic here
//     return failures
// }
//
// Tips for implementing rules:
// - Use file.AST to access the abstract syntax tree
// - Walk the AST to find nodes matching your criteria
// - Create lint.Failure objects for violations found
// - Use appropriate severity levels (error, warning, etc.)
// - Leverage golang.org/x/tools/go/ast packages for AST manipulation
