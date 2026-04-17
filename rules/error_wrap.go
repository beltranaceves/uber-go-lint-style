package rules

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ErrorWrapRule flags fmt.Errorf calls that start error messages with
// "failed to" and suggests using a concise context instead.
type ErrorWrapRule struct{}

func (r *ErrorWrapRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "error_wrap",
		Doc: `avoid using the "failed to" prefix when adding context to errors.

This rule detects calls to fmt.Errorf where the format string begins with
"failed to" (or "failed") and reports a diagnostic suggesting a more
succinct context (for example: "new store: %w").`,
		Run: r.run,
	}
}

func (r *ErrorWrapRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			// Expecting fmt.Errorf
			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if ident.Name != "fmt" || sel.Sel.Name != "Errorf" {
				return true
			}
			if len(call.Args) == 0 {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			s, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			lower := strings.ToLower(s)
			if strings.HasPrefix(lower, "failed to ") || strings.HasPrefix(lower, "failed ") {
				pass.Report(analysis.Diagnostic{
					Pos:     lit.Pos(),
					Message: "avoid 'failed to' prefix in error messages; use concise context, e.g. new store: %w",
				})
			}
			return true
		})
	}
	return nil, nil
}
