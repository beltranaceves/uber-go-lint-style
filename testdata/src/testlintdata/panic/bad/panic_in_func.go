package panicbad

func Do() {
	// BAD: explicit panic in a normal function
	panic("unrecoverable") // want "avoid using panic for error handling; return an error instead"
}
