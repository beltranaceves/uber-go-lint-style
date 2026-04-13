package functional_option

// Test intent: edge cases for parameters counting and methods

type S struct{}

// BAD: exported method (name exported) with 3 parameters should trigger
func (S) Do(a, b, c int) { // want "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments"
}

// BAD: exported function with variadic plus other params (variadic counts as one field but names may be multiple)
func WithVariadic(a int, b string, others ...int) {} // want "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments"

// GOOD: unexported function should not be flagged even with 3+ params
func helper(a, b, c int) {}
