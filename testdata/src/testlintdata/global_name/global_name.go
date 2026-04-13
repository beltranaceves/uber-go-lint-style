package global_name

import "errors"

// BAD: const without underscore
const (
	defaultPort = 8080 // want "unexported package-level identifier 'defaultPort' should be prefixed with '_'"
	_defaultOK  = 9090
)

// BAD: var without underscore
var (
	defaultUser = "user" // want "unexported package-level identifier 'defaultUser' should be prefixed with '_'"
	_allowed    = "ok"
)

// GOOD: exported names are allowed
var DefaultTimeout = 5

// GOOD: unexported error sentinel may use 'err' prefix
var errSomething = errors.New("boom")

// LOCAL: function-local variables should not be flagged
func local() {
	defaultLocal := 1
	_ = defaultLocal
}
