package var_decl

// BAD: prefer := for local var with initializer
func bad1() {
	var s = "foo" // want "prefer ':=' for local variable declarations with initializers"
	_ = s
}

// BAD: multiple names with values
func bad2() {
	var a, b = 1, 2 // want "prefer ':=' for local variable declarations with initializers"
	_ = a + b
}

// GOOD: explicit type (allowed)
func good1() {
	var s string = "foo"
	_ = s
}

// GOOD: empty slice declaration (allowed)
func good2() {
	var filtered []int
	_ = filtered
}
