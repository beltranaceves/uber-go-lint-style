package global_decl

// Test intent: mixed BAD/GOOD examples for global_decl rule

// BAD: explicit type matches initializer (function return)
var AVal int

func AValf() int { return 1 }

var a int = AValf() // want "omit the explicit type in top-level var; use var name = expr instead"

// BAD: multiple names with corresponding values -> two diagnostics
var a1, a2 int = AValf(), AValf() // want "omit the explicit type in top-level var; use var name = expr instead"

// GOOD: no initializer -> explicit type may be needed
var explicitOnly int

// GOOD: explicit type differs from initializer (widening to interface)
type myError2 struct{}

func (myError2) Error() string { return "" }
func ReturnMyError2() myError2 { return myError2{} }

var errVal error = ReturnMyError2()
