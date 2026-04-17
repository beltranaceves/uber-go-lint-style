package printf_name

// BAD: missing trailing 'f'
func Wrap(format string, a ...interface{}) { // want "printf-style function 'Wrap' should be named 'Wrapf'"
}

// GOOD
func Wrapf(format string, a ...interface{}) {
}

type S struct{}

// BAD: method should end with 'f'
func (s *S) Log(format string, args ...interface{}) { // want "printf-style function 'Log' should be named 'Logf'"
}

// NOT A PRINTF: variadic of strings should not trigger
func Join(sep string, parts ...string) {
}
