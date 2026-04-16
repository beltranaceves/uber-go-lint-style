package rules

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoPanicRule flags use of the built-in panic() call outside of init functions.
type NoPanicRule struct{}

// BuildAnalyzer returns the analyzer for the panic rule.
func (r *NoPanicRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "panic",
		Doc: `avoid using panic for error handling or control flow.

This rule reports explicit calls to the built-in panic() function. Panics
are allowed only during program initialization (for example inside an
init() function). In tests prefer t.Fatal/t.FailNow over panics. Use
//nolint:panic to suppress when an explicit panic is intentional.`,
		Run: r.run,
	}
}

func (r *NoPanicRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// Inspect top-level nodes. When encountering a FuncDecl we inspect
		// its body separately so we know the enclosing function name.
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				fname := ""
				if node.Name != nil {
					fname = node.Name.Name
				}
				ast.Inspect(node.Body, func(m ast.Node) bool {
					if call, ok := m.(*ast.CallExpr); ok {
						if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
							if fname != "init" {
								pass.Report(analysis.Diagnostic{
									Pos:     call.Pos(),
									Message: "avoid using panic for error handling; return an error instead",
								})
							}
						}
					}
					return true
				})
				// skip descending into the FuncDecl body again
				return false

			default:
				// Check for top-level explicit panic calls (not in any function)
				if call, ok := n.(*ast.CallExpr); ok {
					if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
						pass.Report(analysis.Diagnostic{
							Pos:     call.Pos(),
							Message: "avoid using panic for error handling; return an error instead",
						})
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
