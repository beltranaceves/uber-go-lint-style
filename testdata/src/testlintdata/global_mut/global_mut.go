package globalmut

import (
	"errors"
	"time"
)

// BAD: package-level mutable function value
var _timeNow = time.Now // want "avoid mutable package-level variable '_timeNow'; prefer dependency injection or scoped state"

// BAD: simple mutable package-level variable
var counter = 0 // want "avoid mutable package-level variable 'counter'; prefer dependency injection or scoped state"

// GOOD: const is allowed
const Version = "1.0"

// GOOD: sentinel error is allowed
var ErrNotFound = errors.New("not found")

// GOOD: exported package API often uses package-level values
var ExportedCounter = 0

// Mixed spec: first is literal (skip), second is call -> report only 'b'
func foo() int { return 1 }

var a, b = 1, foo() // want "avoid mutable package-level variables 'a, b'; prefer dependency injection or scoped state"

// Multi-name spec where both are flagged -> single diagnostic listing both
func f() int { return 1 }
func g() int { return 2 }

var x, y = f(), g() // want "avoid mutable package-level variables 'x, y'; prefer dependency injection or scoped state"

// GOOD: scoped variable inside a function
func UseLocal() int {
	var local = 1
	return local
}

// GOOD: state hidden behind a type
type holder struct {
	n int
}

func newHolder() *holder { return &holder{n: 0} }
