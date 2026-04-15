package interface_pointer

import (
	"bytes"
	"io"
)

// BAD: pointer to named interface in parameter
func BadParam(r *io.Reader) { // want "pointer to interface is unnecessary; pass the interface value instead"
    _ = r
}

// BAD: pointer to interface field
type BadStruct struct {
    R *io.Reader // want "pointer to interface is unnecessary; pass the interface value instead"
}

// BAD: var declaration
var GlobalReader *io.Reader // want "pointer to interface is unnecessary; pass the interface value instead"

// GOOD: interface value
func GoodParam(r io.Reader) {
    _ = r
}

type GoodStruct struct {
    R io.Reader
}

// GOOD: pointer to concrete type that implements io.Reader
var Buf *bytes.Buffer
