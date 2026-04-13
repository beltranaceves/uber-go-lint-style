package functional_option

// Test intent: multiple violations in one file

func Abad(a, b, c int) {} // want "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments"

func Bbad(a int, b int, c int) {} // want "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments"

// GOOD: small functions OK
func small(a int) {}
