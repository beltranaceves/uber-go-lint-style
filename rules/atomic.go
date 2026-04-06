package rules

import (
	"go/ast"
	"strconv"

	"github.com/mgechev/revive/lint"
)

// AtomicRule lints atomic according to Uber Go Style Guide.
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
			Confidence: 1,
		})
	}

	ast.Inspect(file.AST, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		pkgIdent, ok := sel.X.(*ast.Ident)
		if !ok || pkgIdent.Obj != nil {
			return true
		}

		for _, imp := range file.AST.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil || path != "sync/atomic" {
				continue
			}

			name := "atomic"
			if imp.Name != nil {
				name = imp.Name.Name
			}

			if name == "." || name == "_" {
				continue
			}

			if pkgIdent.Name == name {
				failures = append(failures, lint.Failure{
					Failure:    "Use go.uber.org/atomic wrappers rather than sync/atomic operations.",
					Node:       sel,
					Confidence: 1,
				})
				break
			}
		}

		return true
	})

	return failures
}
