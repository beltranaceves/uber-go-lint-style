package panicgood

func init() {
	// Allowed: anonymous function that panics during init is permitted
	func() {
		panic("init panic allowed")
	}()

	// Also allowed when panicking inside an anonymous function started from init
	go func() {
		panic("init goroutine panic allowed")
	}()
}
