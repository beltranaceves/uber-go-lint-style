package else_unnecessary

// BAD: else is unnecessary when both branches assign the same variable
func bad(b bool) {
	var a int
	if b { // want "else is unnecessary: both branches assign the same variable; initialize before the if and remove the else"
		a = 100
	} else {
		a = 10
	}
	_ = a
}

// GOOD: initialize and keep a single if branch
func good(b bool) {
	a := 10
	if b {
		a = 100
	}
	_ = a
}
