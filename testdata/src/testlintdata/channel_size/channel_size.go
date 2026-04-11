package channel_size

// BAD: buffered with a large size
func badLiteral() {
	c := make(chan int, 64) // want "channel size should be one or unbuffered"
	_ = c
}

// BAD: dynamic size
func badDynamic() {
	n := 4
	c := make(chan int, n) // want "channel size should be one or unbuffered"
	_ = c
}

// GOOD: dynamic size, but ignored linter
// nolint: channel_size
func goodDynamic() {
	n := 4
	c := make(chan int, n)
	_ = c
}

// GOOD: size of one
func goodOne() {
	c := make(chan int, 1)
	_ = c
}

// GOOD: unbuffered
func goodUnbuffered() {
	c := make(chan int)
	_ = c
}
