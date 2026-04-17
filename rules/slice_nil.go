package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// SliceNilRule enforces using nil for zero-length slices and using len(...) to
// check emptiness.
type SliceNilRule struct{}

// BuildAnalyzer returns the analyzer for the slice_nil rule
func (r *SliceNilRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "slice_nil",
		Doc: `prefer nil for zero-length slices and use len(s) to check emptiness.

This rule detects returning or initializing an empty slice literal (e.g. []T{} or make([]T, 0))
and comparisons of a slice to nil when the intent is to check emptiness.
Use ` + "`nil`" + ` for zero-length slices and ` + "`len(s) == 0`" + ` to check emptiness.
`,
		Run: r.run,
	}
}

func isEmptyCompositeLit(cl *ast.CompositeLit) bool {
	if cl == nil {
		return false
	}
	// Must be a slice composite literal like []T{}
	if _, ok := cl.Type.(*ast.ArrayType); !ok {
		return false
	}
	return len(cl.Elts) == 0
}

func isMakeEmpty(call *ast.CallExpr) bool {
	if call == nil {
		return false
	}
	if id, ok := call.Fun.(*ast.Ident); ok {
		if id.Name != "make" {
			return false
		}
		// make([]T, 0, ...) -> second arg literal 0
		if len(call.Args) >= 2 {
			if bl, ok := call.Args[1].(*ast.BasicLit); ok {
				return bl.Value == "0"
			}
		}
	}
	return false
}

func isNilIdent(expr ast.Expr) bool {
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name == "nil"
	}
	return false
}

func isSliceType(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Slice)
	return ok
}

func (r *SliceNilRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.ReturnStmt:
				for _, res := range node.Results {
					// return []T{}
					if cl, ok := res.(*ast.CompositeLit); ok {
						if isEmptyCompositeLit(cl) {
							pass.Report(analysis.Diagnostic{
								Pos:     cl.Pos(),
								Message: "prefer returning nil for zero-length slices",
							})
						}
					}
					// return make([]T, 0)
					if call, ok := res.(*ast.CallExpr); ok {
						if isMakeEmpty(call) {
							pass.Report(analysis.Diagnostic{
								Pos:     call.Pos(),
								Message: "prefer returning nil for zero-length slices",
							})
						}
					}
				}

			case *ast.ValueSpec:
				for i, val := range node.Values {
					// var x = []T{}
					if cl, ok := val.(*ast.CompositeLit); ok {
						if isEmptyCompositeLit(cl) {
							pos := cl.Pos()
							// If the name exists, prefer pointing at it when present
							if i < len(node.Names) {
								pos = node.Names[i].Pos()
							}
							pass.Report(analysis.Diagnostic{
								Pos:     pos,
								Message: "prefer nil slice for zero-value slice declarations",
							})
						}
					}
					if call, ok := val.(*ast.CallExpr); ok {
						if isMakeEmpty(call) {
							pos := call.Pos()
							if i < len(node.Names) {
								pos = node.Names[i].Pos()
							}
							pass.Report(analysis.Diagnostic{
								Pos:     pos,
								Message: "prefer nil slice for zero-value slice declarations",
							})
						}
					}
				}

			case *ast.AssignStmt:
				for i, rh := range node.Rhs {
					if cl, ok := rh.(*ast.CompositeLit); ok {
						if isEmptyCompositeLit(cl) {
							pos := cl.Pos()
							// Try to point to LHS name if available
							if i < len(node.Lhs) {
								pos = node.Lhs[i].Pos()
							}
							pass.Report(analysis.Diagnostic{
								Pos:     pos,
								Message: "prefer nil slice for zero-value slice declarations",
							})
						}
					}
					if call, ok := rh.(*ast.CallExpr); ok {
						if isMakeEmpty(call) {
							pos := call.Pos()
							if i < len(node.Lhs) {
								pos = node.Lhs[i].Pos()
							}
							pass.Report(analysis.Diagnostic{
								Pos:     pos,
								Message: "prefer nil slice for zero-value slice declarations",
							})
						}
					}
				}

			case *ast.BinaryExpr:
				if node.Op == token.EQL || node.Op == token.NEQ {
					// x == nil  or nil == x
					var leftIsNil = isNilIdent(node.X)
					var rightIsNil = isNilIdent(node.Y)

					if leftIsNil && !rightIsNil {
						if isSliceType(pass.TypesInfo.TypeOf(node.Y)) {
							pass.Report(analysis.Diagnostic{
								Pos:     node.Pos(),
								Message: "use len(s) == 0 to check for empty slices",
							})
						}
					} else if rightIsNil && !leftIsNil {
						if isSliceType(pass.TypesInfo.TypeOf(node.X)) {
							pass.Report(analysis.Diagnostic{
								Pos:     node.Pos(),
								Message: "use len(s) == 0 to check for empty slices",
							})
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}
