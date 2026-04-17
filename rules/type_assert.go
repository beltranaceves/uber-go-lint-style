package rules

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// TypeAssertRule enforces using the comma-ok idiom for type assertions
type TypeAssertRule struct{}

type typeAssertWalker struct {
	pass  *analysis.Pass
	stack []ast.Node
}

func (w *typeAssertWalker) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		if len(w.stack) > 0 {
			w.stack = w.stack[:len(w.stack)-1]
		}
		return nil
	}

	w.stack = append(w.stack, n)

	if ta, ok := n.(*ast.TypeAssertExpr); ok {
		var p ast.Node
		if len(w.stack) >= 2 {
			p = w.stack[len(w.stack)-2]
		}
		if as, ok2 := p.(*ast.AssignStmt); !(ok2 && len(as.Lhs) >= 2) {
			w.pass.Report(analysis.Diagnostic{
				Pos:      ta.Pos(),
				End:      0,
				Category: "type_assert",
				Message:  "use the comma-ok form for type assertions",
			})
		}
	}

	return w
}

// BuildAnalyzer returns the analyzer for the type_assert rule
func (r *TypeAssertRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "type_assert",
		Doc:  "always use the comma-ok form for type assertions to avoid panics",
		Run:  r.run,
	}
}

func (r *TypeAssertRule) run(pass *analysis.Pass) (any, error) {
	w := &typeAssertWalker{pass: pass}

	for _, file := range pass.Files {
		ast.Walk(w, file)
	}

	return nil, nil
}
