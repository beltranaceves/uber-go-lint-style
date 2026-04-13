package global_decl

// BAD: explicit type is redundant when initializer provides same type
var _s string = F() // want "omit the explicit type in top-level var; use var name = expr instead"

func F() string { return "A" }

// GOOD: type is inferred
var _s2 = F()

// GOOD: declare type when initializer's type differs
type myError struct{}

func (myError) Error() string { return "error" }

func G() myError { return myError{} }

var _e error = G()
