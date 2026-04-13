package error_name

import (
	"errors"
	"fmt"
)

// BAD: exported error variable must start with Err
var BrokenLink = errors.New("broken") // want "exported error variable should be prefixed with 'Err'"

// BAD: unexported error variable must start with err
var notFound = fmt.Errorf("not found") // want "unexported error variable should be prefixed with 'err'"

// GOOD: exported and unexported error variables with correct prefixes
var ErrCouldNotOpen = errors.New("could not open")
var errInternal = errors.New("internal")

// BAD: type implements Error but name doesn't end with Error
type NotFound struct{} // want "error type names should end with 'Error'"

func (n NotFound) Error() string {
    return "not found"
}

// GOOD: error type named with Error suffix
type ResolveError struct{}

func (r ResolveError) Error() string {
    return "resolve"
}
