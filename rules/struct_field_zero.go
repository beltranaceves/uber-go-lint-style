package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"go/types"

	"golang.org/x/tools/go/analysis"
)

// StructFieldZeroRule suggests omitting zero-valued fields in keyed struct literals.
type StructFieldZeroRule struct{}

func (r *StructFieldZeroRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "struct_field_zero",
		Doc: `omit zero-valued fields in keyed struct literals.

This rule reports struct literal fields that are explicitly set to a type's
zero value (for example "", 0, false, or nil) and suggests letting Go
initialize them implicitly. Test tables (variables named tests) are exempt
because named zero-values are often meaningful in test cases.
`,
		Run: r.run,
	}
}

type sfzVisitor struct {
	pass  *analysis.Pass
	stack []ast.Node
}

func (v *sfzVisitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		if len(v.stack) > 0 {
			v.stack = v.stack[:len(v.stack)-1]
		}
		return nil
	}
	v.stack = append(v.stack, n)

	cl, ok := n.(*ast.CompositeLit)
	if !ok {
		return v
	}

	tv, ok := v.pass.TypesInfo.Types[cl]
	if !ok || tv.Type == nil {
		return v
	}
	st, ok := tv.Type.Underlying().(*types.Struct)
	if !ok {
		return v
	}

	// Detect if this literal appears in a test table named `tests`.
	inTests := false
	for i := len(v.stack) - 2; i >= 0; i-- {
		switch node := v.stack[i].(type) {
		case *ast.ValueSpec:
			for _, name := range node.Names {
				if name.Name == "tests" {
					inTests = true
					break
				}
			}
			if inTests {
				break
			}
		case *ast.AssignStmt:
			for _, expr := range node.Lhs {
				if id, ok := expr.(*ast.Ident); ok && id.Name == "tests" {
					inTests = true
					break
				}
			}
			if inTests {
				break
			}
		}
		if inTests {
			break
		}
	}

	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		// key must be identifier (struct field name)
		id, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		// Find field type by name
		var ftype types.Type
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			if f.Name() == id.Name {
				ftype = f.Type()
				break
			}
		}
		if ftype == nil {
			continue
		}

		if isZeroLiteralForType(v.pass, kv.Value, ftype) {
			if inTests {
				continue
			}
			v.pass.Report(analysis.Diagnostic{
				Pos:     kv.Pos(),
				Message: fmt.Sprintf("omit zero-valued field %q from struct literal; let Go set the zero value", id.Name),
			})
		}
	}
	return v
}

func isZeroLiteralForType(pass *analysis.Pass, expr ast.Expr, typ types.Type) bool {
	// Handle nil literal
	if id, ok := expr.(*ast.Ident); ok {
		if id.Name == "nil" {
			// nil is zero for pointer, slice, map, chan, func, interface
			switch typ.Underlying().(type) {
			case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature, *types.Interface:
				return true
			}
		}
		if id.Name == "false" {
			if b, ok := typ.Underlying().(*types.Basic); ok && b.Kind() == types.Bool {
				return true
			}
		}
		return false
	}

	// Basic literals: string, int, float, rune
	if bl, ok := expr.(*ast.BasicLit); ok {
		switch bl.Kind {
		case token.STRING:
			// Value includes quotes, compare to empty string literal
			return bl.Value == "\"\""
		case token.INT:
			// conservative: only flag literal "0"
			if bl.Value == "0" {
				if b, ok := typ.Underlying().(*types.Basic); ok {
					switch b.Kind() {
					case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
						types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
						return true
					}
				}
			}
		case token.FLOAT:
			if bl.Value == "0" || bl.Value == "0.0" {
				if b, ok := typ.Underlying().(*types.Basic); ok {
					switch b.Kind() {
					case types.Float32, types.Float64:
						return true
					}
				}
			}
		}
	}

	// Other expressions are not considered zero-valued by this conservative check.
	_ = pass
	_ = strings.TrimSpace
	return false
}

func (r *StructFieldZeroRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Walk(&sfzVisitor{pass: pass}, file)
	}
	return nil, nil
}
