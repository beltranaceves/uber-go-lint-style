package rules

import (
	"go/ast"

	"go/types"

	"golang.org/x/tools/go/analysis"
)

// InterfacePointerRule flags uses of pointer-to-interface types.
type InterfacePointerRule struct{}

func (r *InterfacePointerRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "interface_pointer",
		Doc: `avoid pointers to interface types.

Pointer-to-interface types (for example, *io.Reader) are unnecessary —
interfaces should be passed by value. If you need methods to mutate the
underlying concrete value, use a pointer receiver on the concrete type,
not a pointer to the interface.
`,
		Run: r.run,
	}
}

func (r *InterfacePointerRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			star, ok := n.(*ast.StarExpr)
			if !ok {
				return true
			}

			// Use type information to determine whether this StarExpr denotes
			// a pointer type whose element is an interface.
			t := pass.TypesInfo.TypeOf(star)
			if t == nil {
				t = pass.TypesInfo.TypeOf(star.X)
			}
			ptr, ok := t.(*types.Pointer)
			if !ok {
				return true
			}

			if _, ok := ptr.Elem().Underlying().(*types.Interface); ok {
				pass.Report(analysis.Diagnostic{
					Pos:     star.Pos(),
					Message: "pointer to interface is unnecessary; pass the interface value instead",
				})
			}

			return true
		})
	}
	return nil, nil
}
