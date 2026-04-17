package struct_zero

type User struct{}

// BAD: short declaration using composite literal
func badShort() {
	user := User{} // want "use var for zero-value structs"
	_ = user
}

// BAD: var with composite literal initializer
func badVarInit() {
	var user = User{} // want "use var for zero-value structs"
	_ = user
}

// GOOD: prefer var without initializer
func goodVar() {
	var user User
	_ = user
}

// GOOD: pointer composite literal should not be flagged
func goodPointer() {
	user := &User{}
	_ = user
}
