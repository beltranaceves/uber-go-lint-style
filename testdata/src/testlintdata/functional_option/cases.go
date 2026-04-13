package functional_option

// Test intent: mixed GOOD/BAD examples for the functional_option rule

// BAD: exported function with 3 explicit params should trigger rule
func OpenBad(addr string, cache bool, logger *int) (*Connection, error) { // want "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments"
	return nil, nil
}

// GOOD: functional options used
func OpenGood(addr string, opts ...Option) (*Connection, error) {
	return nil, nil
}

// BAD: exported function with unnamed parameters (counts each) should trigger
func CreateBad(string, int, int) {} // want "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments"

// GOOD: exported but only two parameters
func CreateGood(a, b int) {}
