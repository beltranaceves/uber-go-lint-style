package builtinname

// BAD EXAMPLE 1: Function parameter shadowing builtin
func example1(error string) { // want "identifier 'error' shadows a built-in, consider using a different name"
	// `error` shadows the builtin
}

// BAD EXAMPLE 2: Function parameter with int
func example2(int int) int { // want "identifier 'int' shadows a built-in, consider using a different name"
	return int + 1
}

// BAD EXAMPLE 3: Struct with shadowing field
type Example3 struct {
	err error // OK - field is "err", not "error"
}

// BAD EXAMPLE 4: Different struct with shadowing field
type Example4 struct {
	error error // want "identifier 'error' shadows a built-in, consider using a different name"
}

// BAD EXAMPLE 5: Receiver parameter shadowing
func (m Example3) String(string string) string { // want "identifier 'string' shadows a built-in, consider using a different name"
	return string
}

// GOOD EXAMPLES - These should NOT trigger lint errors

// Example 6: Good function parameter naming
func example6(msg string) {
	// `error` is not shadowed
}

// Example 7: Good function with int parameter
func example7(val int) int {
	return val + 1
}

// Example 8: Good struct with proper field names
type Example8 struct {
	text  string
	title string
}

// Example 9: Struct correctly using error type
type Example9 struct {
	err error
}

// Example 10: Struct with multiple fields
type Example10 struct {
	name   string
	value  int
	active bool
}

// Example 11: Using builtins correctly (not shadowing)
func example11() {
	// These are OK - we're using builtins correctly, not shadowing them
	_ = len("hello") // OK
	panic("error")   // may not return, but OK - "error" is a builtin but we're calling panic
	var x int        // OK - "int" is only the type, not shadowed
	_ = x
}

// BAD EXAMPLE 12: bool parameter shadowing
func example12(bool bool) bool { // want "identifier 'bool' shadows a built-in, consider using a different name"
	return bool
}

// BAD EXAMPLE 13: Another receiver parameter
type Example13 struct {
	value int
}

func (e Example13) Method(len int) int { // want "identifier 'len' shadows a built-in, consider using a different name"
	return len
}
