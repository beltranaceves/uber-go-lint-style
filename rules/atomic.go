package rules

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// AtomicRule checks for usage of sync/atomic operations on raw types.
type AtomicRule struct{}

// BuildAnalyzer returns the analyzer for the atomic rule
func (r *AtomicRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "atomic",
		Doc: `require use of go.uber.org/atomic instead of sync/atomic for operations on raw types.

			Detects calls to sync/atomic functions that take or return raw types (int32, int64, uint32, uint64, uintptr).
			Raw types lack type safety and are error-prone compared to go.uber.org/atomic's wrapped types.`,
		Run: r.run,
	}
}

// isRawType checks if a type is a raw integer/pointer type that requires go.uber.org/atomic.
// Raw types are: int32, int64, uint32, uint64, uintptr
//
// TODO: Consider improving this by dynamically checking if a type has corresponding
// functions in sync/atomic or go.uber.org/atomic packages, rather than maintaining
// a hardcoded list of raw types. This would make the check more maintainable and
// forward-compatible with potential future atomic types.
func isRawType(t types.Type) bool {
	if t == nil {
		return false
	}

	switch u := t.Underlying().(type) {
	case *types.Basic:
		kind := u.Kind()
		// Check for raw integer types
		return kind == types.Int32 || kind == types.Int64 ||
			kind == types.Uint32 || kind == types.Uint64 ||
			kind == types.Uintptr
	case *types.Pointer:
		// Check for pointers to raw types
		if basic, ok := u.Elem().Underlying().(*types.Basic); ok {
			kind := basic.Kind()
			return kind == types.Int32 || kind == types.Int64 ||
				kind == types.Uint32 || kind == types.Uint64 ||
				kind == types.Uintptr
		}
	}
	return false
}

// functionInvolvesRawType checks if a function signature involves raw types.
// Returns true if any parameter or return value is a raw type.
func functionInvolvesRawType(fn *types.Func) bool {
	if fn == nil {
		return false
	}

	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return false
	}

	// Check parameters
	if sig.Params() != nil {
		for i := 0; i < sig.Params().Len(); i++ {
			if isRawType(sig.Params().At(i).Type()) {
				return true
			}
		}
	}

	// Check results
	if sig.Results() != nil {
		for i := 0; i < sig.Results().Len(); i++ {
			if isRawType(sig.Results().At(i).Type()) {
				return true
			}
		}
	}

	return false
}

func (r *AtomicRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Look for function calls
			if callExpr, ok := n.(*ast.CallExpr); ok {
				// Check if this is a selector expression (e.g., atomic.SomeFunc)
				if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					// Get the package identifier
					if ident, ok := selectorExpr.X.(*ast.Ident); ok && ident.Name == "atomic" {
						// Look up the function in TypesInfo
						if callObj, ok := pass.TypesInfo.Uses[selectorExpr.Sel]; ok {
							if fn, ok := callObj.(*types.Func); ok {
								// Check if the function is from sync/atomic and involves raw types
								if fn.Pkg() != nil && fn.Pkg().Path() == "sync/atomic" {
									if functionInvolvesRawType(fn) {
										pass.Report(analysis.Diagnostic{
											Pos:     callExpr.Pos(),
											Message: "use go.uber.org/atomic instead of sync/atomic for operations on raw types",
										})
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}

	return nil, nil
}
