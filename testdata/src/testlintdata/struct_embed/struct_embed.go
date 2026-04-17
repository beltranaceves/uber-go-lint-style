package struct_embed

import (
	"io"
	"sync"
)

// BAD: embedded after a regular field
type BadAfterRegular struct {
	x         int
	io.Reader // want "embedded field should be placed at the top of the struct"
}

// BAD: no empty line between embedded fields and regular fields
type BadNoEmptyLine struct {
	io.Reader
	x int // want "add an empty line between embedded fields and regular fields"
}

// BAD: embedding sync.Mutex
type BadEmbedMutex struct {
	sync.Mutex // want "do not embed sync.Mutex; use a named field instead"
}

// GOOD: embedded fields at top with empty line
type Good struct {
	io.Reader

	x int
}
