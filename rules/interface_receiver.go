package rules

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// InterfaceReceiverRule warns when code takes a method value (e.g., f := x.M)
// because the receiver is captured at evaluation time. Subsequent mutations to
// the original value or the pointee will not affect the stored receiver.
type InterfaceReceiverRule struct{}

// BuildAnalyzer returns the analyzer for the interface_receiver rule.
func (r *InterfaceReceiverRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "interface_receiver",
		Doc: `warn when a method value is created (for example, f := x.M).

Taking a method value evaluates and saves the receiver at the time the
method value is formed. Subsequent mutations to the original value or the
pointee (for pointer receivers) do not affect the saved receiver. This
analyzer surfaces these sites so the developer can consider whether a
closure or explicit function literal is more appropriate.`,
		Run: r.run,
	}
}

func (r *InterfaceReceiverRule) run(pass *analysis.Pass) (any, error) {
	if pass == nil || pass.TypesInfo == nil {
		return nil, nil
	}

	for _, file := range pass.Files {
		var stack []ast.Node
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
				return true
			}

			var parent ast.Node
			if len(stack) > 0 {
				parent = stack[len(stack)-1]
			}

			if selExpr, ok := n.(*ast.SelectorExpr); ok {
				// If this selector is the function part of a call (x.M()), skip it.
				if call, ok := parent.(*ast.CallExpr); ok && call.Fun == selExpr {
					// method call, nothing to do
				} else {
					if sel := pass.TypesInfo.Selections[selExpr]; sel != nil {
						// We only care about method values (not field selectors,
						// method expressions like T.M, or other selection kinds).
						if sel.Kind() == types.MethodVal {
							if fn, ok := sel.Obj().(*types.Func); ok {
								if sig, ok := fn.Type().(*types.Signature); ok {
									recv := sig.Recv()
									if recv != nil {
										_, isPtr := recv.Type().(*types.Pointer)
										if isPtr {
											pass.Report(analysis.Diagnostic{
												Pos:     selExpr.Pos(),
												Message: "taking a method value with a pointer receiver captures the receiver; subsequent mutations to the pointee will not affect the stored receiver",
											})
										} else {
											pass.Report(analysis.Diagnostic{
												Pos:     selExpr.Pos(),
												Message: "taking a method value captures the receiver by value; subsequent mutations to the original value will not affect the stored receiver",
											})
										}
									}
								}
							}
						}
					}
				}
			}

			stack = append(stack, n)
			return true
		})
	}

	return nil, nil
}
