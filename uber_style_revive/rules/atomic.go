package rules

import (
	"strconv"

	"github.com/mgechev/revive/lint"
)

// AtomicRule enforces using go.uber.org/atomic instead of sync/atomic.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#atomic
type AtomicRule struct{}

// Name returns the rule name
func (r *AtomicRule) Name() string {
	return "atomic"
}

// Apply runs the rule against the provided file
func (r *AtomicRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	if file == nil || file.AST == nil {
		return failures
	}

	// Check for sync/atomic imports
	for _, imp := range file.AST.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}

		if path != "sync/atomic" {
			continue
		}

		failures = append(failures, lint.Failure{
			Failure:    "Prefer go.uber.org/atomic instead of sync/atomic to avoid raw atomic primitives.",
			Node:       imp,
			Confidence: 1.0,
		})
	}

	return failures
}
