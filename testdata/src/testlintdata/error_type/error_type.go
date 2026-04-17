package error_type

import (
	"errors"
	"fmt"
)

// BAD: exported var created with fmt.Errorf
var ErrBad = fmt.Errorf("file %q not found", "name") // want "exported error variable is created by fmt.Errorf"

// OK: exported static error
var ErrGood = errors.New("could not open")

// OK: exported custom error type
type NotFoundError struct{ File string }

func (e *NotFoundError) Error() string { return "not found" }

var ErrCustom = &NotFoundError{File: "x"}

// OK: non-exported var using fmt.Errorf is allowed
var errLocal = fmt.Errorf("local %s", "x")
