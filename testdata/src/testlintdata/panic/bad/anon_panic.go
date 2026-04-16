package panicbad

func DoAnon() {
	// BAD: panic inside anonymous function invoked in a normal function
	func() {
		panic("whoops") // want "avoid using panic for error handling; return an error instead"
	}()
}

func DoGoAnon() {
	// BAD: panic inside goroutine started from a normal function
	go func() {
		panic("boom") // want "avoid using panic for error handling; return an error instead"
	}()
}
