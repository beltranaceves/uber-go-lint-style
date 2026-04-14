package rules

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// MapInitRule prefers make(...) for empty maps and map literals for fixed maps.
type MapInitRule struct{}

func (r *MapInitRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "map_init",
		Doc: `prefer make() for empty maps and map literals for fixed maps.

This rule flags empty map composite literals like map[T]U{} and
recommends using make(map[T]U) instead. Use map literals when initializing a
map with a fixed set of elements.`,
		Run: r.run,
	}
}

func (r *MapInitRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			cl, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}

			// Only care about map composite literals with no elements
			if _, ok := cl.Type.(*ast.MapType); ok && len(cl.Elts) == 0 {
				msg := "use make for empty maps"
				pass.Report(analysis.Diagnostic{Pos: cl.Pos(), Message: msg})
			}

			return true
		})
	}
	return nil, nil
}

// exprString returns a short source-like representation for common node types.
// Keep it minimal to produce readable diagnostic messages.
func shortExpr(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", shortExpr(t.Key), shortExpr(t.Value))
	case *ast.Ident:
		return t.Name
	case *ast.CompositeLit:
		if mt, ok := t.Type.(*ast.MapType); ok {
			return fmt.Sprintf("%s{}", shortExpr(mt))
		}
		return "composite{}"
	default:
		return "map[...]..."
	}
}
