package else_unnecessary

// Selector LHS
func selectorBad(b bool, p *struct{ x int }) {
	var _ = 0
	if b { // want "else is unnecessary: both branches assign the same variable; initialize before the if and remove the else"
		p.x = 100
	} else {
		p.x = 10
	}
}

// Return-only branches
func returnBad(b bool) int {
	if b { // want "else is unnecessary: both branches return; simplify to if-return and a single return"
		return 1
	} else {
		return 2
	}
}

// Side-effecting RHS should NOT be flagged
func sideEffect(b bool) {
	var a int
	if b {
		a = f()
	} else {
		a = 2
	}
	_ = a
}
