package rules

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// DeferCleanRule recommends using `defer` to clean up resources such as files and locks.
type DeferCleanRule struct{}

// BuildAnalyzer returns the analyzer for the defer-clean rule
func (r *DeferCleanRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "defer_clean",
		Doc: `use defer to clean up resources such as files and locks.

This rule reports explicit calls to common cleanup methods (for example
` + "`Unlock`" + ` and ` + "`Close`" + `) that are not wrapped in a defer.  Using
defer immediately after acquiring a resource reduces the chance of missing
cleanup across early returns and makes code easier to reason about.

Notes: This rule is conservative and may produce false positives in cases
where manual cleanup is intentional (for example, unlocking inside loops).
`,
		Run: r.run,
	}
}

func (r *DeferCleanRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// collect positions of deferred cleanup calls so we don't report them
		deferred := make(map[token.Pos]bool)

		ast.Inspect(file, func(n ast.Node) bool {
			if ds, ok := n.(*ast.DeferStmt); ok {
				call := ds.Call
				if call != nil {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						switch sel.Sel.Name {
						case "Unlock", "RUnlock", "Close":
							deferred[call.Pos()] = true
						}
					}
				}
			}
			return true
		})

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			switch sel.Sel.Name {
			case "Unlock", "RUnlock", "Close":
				if !deferred[call.Pos()] {
					pass.Report(analysis.Diagnostic{
						Pos:      call.Pos(),
						End:      call.End(),
						Category: "defer-clean",
						Message:  "use defer to clean up resources such as files and locks",
					})
				}
			}

			return true
		})
	}

	return nil, nil
}
