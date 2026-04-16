package rules

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NestLessRule warns when code contains excessive nesting levels.
type NestLessRule struct {
	// MaxDepth sets the maximum allowed nesting. If 0, defaults to 3.
	MaxDepth int
}

func (r *NestLessRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "nest_less",
		Doc: `reduce nesting levels in functions by handling error/special cases early

This rule reports when a block is nested more than a small threshold (default 3).
Deeply nested code is harder to read; prefer early returns or continue statements
to reduce indentation depth.
`,
		Run: r.run,
	}
}

func (r *NestLessRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// Inspect top-level declarations to find function bodies
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			// Determine depth limit for this rule (default 3)
			depthLimit := r.MaxDepth
			if depthLimit == 0 {
				depthLimit = 3
			}

			// Walk statements in function body
			var walkStmt func(ast.Stmt, int)
			walkStmt = func(s ast.Stmt, depth int) {
				switch x := s.(type) {
				case *ast.BlockStmt:
					for _, stmt := range x.List {
						walkStmt(stmt, depth)
					}
				case *ast.IfStmt:
					newDepth := depth + 1
					if newDepth > depthLimit {
						pass.Report(analysis.Diagnostic{
							Pos:     x.Pos(),
							End:     x.End(),
							Message: fmt.Sprintf("reduce nesting: depth %d exceeds allowed %d; consider returning early", newDepth, depthLimit),
						})
					}
					if x.Body != nil {
						walkStmt(x.Body, newDepth)
					}
					if x.Else != nil {
						if elseIf, ok := x.Else.(*ast.IfStmt); ok {
							walkStmt(elseIf, newDepth)
						} else if elseBlock, ok := x.Else.(*ast.BlockStmt); ok {
							walkStmt(elseBlock, newDepth)
						}
					}
				case *ast.ForStmt:
					newDepth := depth + 1
					if newDepth > depthLimit {
						pass.Report(analysis.Diagnostic{
							Pos:     x.Pos(),
							End:     x.End(),
							Message: fmt.Sprintf("reduce nesting: depth %d exceeds allowed %d; consider returning early", newDepth, depthLimit),
						})
					}
					if x.Body != nil {
						walkStmt(x.Body, newDepth)
					}
				case *ast.RangeStmt:
					newDepth := depth + 1
					if newDepth > depthLimit {
						pass.Report(analysis.Diagnostic{
							Pos:     x.Pos(),
							End:     x.End(),
							Message: fmt.Sprintf("reduce nesting: depth %d exceeds allowed %d; consider returning early", newDepth, depthLimit),
						})
					}
					if x.Body != nil {
						walkStmt(x.Body, newDepth)
					}
				case *ast.SwitchStmt:
					newDepth := depth + 1
					if newDepth > depthLimit {
						pass.Report(analysis.Diagnostic{
							Pos:     x.Pos(),
							End:     x.End(),
							Message: fmt.Sprintf("reduce nesting: depth %d exceeds allowed %d; consider returning early", newDepth, depthLimit),
						})
					}
					if x.Body != nil {
						walkStmt(x.Body, newDepth)
					}
				case *ast.TypeSwitchStmt:
					newDepth := depth + 1
					if newDepth > depthLimit {
						pass.Report(analysis.Diagnostic{
							Pos:     x.Pos(),
							End:     x.End(),
							Message: fmt.Sprintf("reduce nesting: depth %d exceeds allowed %d; consider returning early", newDepth, depthLimit),
						})
					}
					if x.Body != nil {
						walkStmt(x.Body, newDepth)
					}
				case *ast.SelectStmt:
					newDepth := depth + 1
					if newDepth > depthLimit {
						pass.Report(analysis.Diagnostic{
							Pos:     x.Pos(),
							End:     x.End(),
							Message: fmt.Sprintf("reduce nesting: depth %d exceeds allowed %d; consider returning early", newDepth, depthLimit),
						})
					}
					if x.Body != nil {
						walkStmt(x.Body, newDepth)
					}
				case *ast.CaseClause:
					for _, stmt := range x.Body {
						walkStmt(stmt, depth+1)
					}
				default:
					// For other statement types, do nothing
				}
			}

			// Start at depth 0 for top-level statements
			for _, stmt := range fn.Body.List {
				walkStmt(stmt, 0)
			}
		}
	}
	return nil, nil
}
