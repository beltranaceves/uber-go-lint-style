package functional_option

// Test intent: parameters are used before any return, so should NOT be flagged

// All parameters are referenced before the function returns.
func UseBeforeReturn(a, b, c int) int {
	_ = a
	_ = b
	_ = c
	return 0
}

// Another case: multiple returns but uses happen before returns
func UseBeforeMultipleReturns(a, b, c int) int { // want "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments"
	_ = a
	if a > 0 {
		_ = b
		return 1
	}
	_ = c
	return 2
}
