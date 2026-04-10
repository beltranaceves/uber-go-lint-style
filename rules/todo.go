package rules

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// TodoRule finds TODO comments without an author
type TodoRule struct{}

// BuildAnalyzer returns the analyzer for the todo rule
func (r *TodoRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "todo",
		Doc:  "find todos without an author",
		Run:  r.run,
	}
}

func (r *TodoRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if comment, ok := n.(*ast.Comment); ok {
				if strings.HasPrefix(comment.Text, "// TODO:") || strings.HasPrefix(comment.Text, "// TODO():") {
					pass.Report(analysis.Diagnostic{
						Pos:            comment.Pos(),
						End:            0,
						Category:       "todo",
						Message:        "TODO comment has no author",
						SuggestedFixes: nil,
					})
				}
			}

			return true
		})
	}

	return nil, nil
}
