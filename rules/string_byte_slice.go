package rules

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// StringByteSliceRule detects repeated conversions of string literals to []byte
// inside loops. Instead of converting inside the loop, convert once and reuse
// the result.
type StringByteSliceRule struct{}

func (r *StringByteSliceRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "string_byte_slice",
		Doc: `avoid repeated string-to-byte conversions.

This rule detects conversions of a string literal to []byte that occur inside
loops. Performing the conversion once outside the loop and reusing the result
is more efficient.`,
		Run: r.run,
	}
}

func isByteArrayType(expr ast.Expr) bool {
	at, ok := expr.(*ast.ArrayType)
	if !ok {
		return false
	}
	if ident, ok := at.Elt.(*ast.Ident); ok {
		return ident.Name == "byte"
	}
	return false
}

func (r *StringByteSliceRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		v := &visitor{pass: pass}
		ast.Walk(v, file)
	}
	return nil, nil
}

type visitor struct {
	pass  *analysis.Pass
	stack []ast.Node
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		if len(v.stack) > 0 {
			v.stack = v.stack[:len(v.stack)-1]
		}
		return nil
	}

	v.stack = append(v.stack, n)

	if call, ok := n.(*ast.CallExpr); ok {
		// Check for conversion of the form: []byte("literal")
		if isByteArrayType(call.Fun) && len(call.Args) == 1 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				// If any ancestor is a for/range loop, report
				for _, anc := range v.stack[:len(v.stack)-1] {
					switch anc.(type) {
					case *ast.ForStmt, *ast.RangeStmt:
						v.pass.Report(analysis.Diagnostic{
							Pos:     call.Pos(),
							Message: "do not convert a string literal to a byte slice repeatedly; convert it once outside the loop and reuse the result",
						})
						// report once per call
						goto done
					}
				}
			}
		}
	}
done:
	return v
}
