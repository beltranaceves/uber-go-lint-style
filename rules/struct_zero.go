package rules

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// StructZeroRule enforces using the var form for zero-value struct declarations.
type StructZeroRule struct{}

func (r *StructZeroRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "struct_zero",
		Doc: `use var for zero-value structs.

When all the fields of a struct are omitted in a declaration, prefer the
` + "`var`" + ` form (e.g. "var u User") instead of a composite literal
initialization (e.g. "u := User{}" or "var u = User{}"). This differentiates
zero-valued structs from those with non-zero fields and matches the style
guideline in the project.
`,
		Run: r.run,
	}
}

func isEmptyStructCompositeLit(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	if cl == nil {
		return false
	}
	if len(cl.Elts) != 0 {
		return false
	}
	// Use type information to ensure this is a struct composite literal.
	t := pass.TypesInfo.TypeOf(cl)
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Struct)
	return ok
}

func (r *StructZeroRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.ReturnStmt:
				for _, res := range node.Results {
					// return User{}
					if cl, ok := res.(*ast.CompositeLit); ok {
						if isEmptyStructCompositeLit(pass, cl) {
							pass.Report(analysis.Diagnostic{Pos: cl.Pos(), Message: "use var for zero-value structs"})
						}
					}
					// return &User{} or other unary expr: don't flag (address creates non-zero pointer)
				}

			case *ast.ValueSpec:
				for i, val := range node.Values {
					// var x = User{}
					if cl, ok := val.(*ast.CompositeLit); ok {
						if isEmptyStructCompositeLit(pass, cl) {
							pos := cl.Pos()
							if i < len(node.Names) {
								pos = node.Names[i].Pos()
							}
							pass.Report(analysis.Diagnostic{Pos: pos, Message: "use var for zero-value structs"})
						}
					}
					// var x = &User{} -> val is UnaryExpr, skip
					if ue, ok := val.(*ast.UnaryExpr); ok {
						if ue.Op.String() == "&" {
							// if the operand is a composite literal, treat as pointer init; do not flag
							// nothing to do
							_ = ue
						}
					}
				}

			case *ast.AssignStmt:
				for i, rh := range node.Rhs {
					if cl, ok := rh.(*ast.CompositeLit); ok {
						if isEmptyStructCompositeLit(pass, cl) {
							pos := cl.Pos()
							if i < len(node.Lhs) {
								pos = node.Lhs[i].Pos()
							}
							pass.Report(analysis.Diagnostic{Pos: pos, Message: "use var for zero-value structs"})
						}
					}
					if ue, ok := rh.(*ast.UnaryExpr); ok {
						// skip &User{} cases
						_ = ue
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
