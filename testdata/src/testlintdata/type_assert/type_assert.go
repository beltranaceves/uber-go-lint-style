package type_assert

// BAD: should use comma-ok form
func bad1() {
	var i interface{} = "x"
	_ = i.(string) // want "use the comma-ok form for type assertions"
}

// GOOD: uses comma-ok
func good1() {
	var i interface{} = "x"
	if s, ok := i.(string); ok {
		_ = s
	}
}
