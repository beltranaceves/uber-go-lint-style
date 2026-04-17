package rules

import (
	"go/ast"

	"go/types"

	"golang.org/x/tools/go/analysis"
)

// StructFieldKeyRule enforces using field names when initializing structs.
type StructFieldKeyRule struct{}

func (r *StructFieldKeyRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "struct_field_key",
		Doc: `require field names when initializing struct composite literals.

This rule reports composite literals that initialize struct types using
positional elements rather than keyed fields (e.g. Field: value). An
exception is made for small test table entries: when the literal is used in
a tests table and the literal has three or fewer fields, positional
initialization is allowed.
`,
		Run: r.run,
	}
}

type structFieldVisitor struct {
	pass  *analysis.Pass
	stack []ast.Node
}

func (v *structFieldVisitor) Visit(n ast.Node) ast.Visitor {
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
	// Only care about struct types
	if _, ok := tv.Type.Underlying().(*types.Struct); !ok {
		return v
	}

	// If any element is a keyed element, it's fine.
	for _, elt := range cl.Elts {
		if _, keyed := elt.(*ast.KeyValueExpr); keyed {
			return v
		}
	}

	// Exception: small test tables named `tests` with <= 3 fields
	allowed := false
	if len(cl.Elts) <= 3 {
		for i := len(v.stack) - 2; i >= 0; i-- {
			switch node := v.stack[i].(type) {
			case *ast.ValueSpec:
				for _, name := range node.Names {
					if name.Name == "tests" {
						allowed = true
						break
					}
				}
				if allowed {
					break
				}
			case *ast.AssignStmt:
				for _, expr := range node.Lhs {
					if id, ok := expr.(*ast.Ident); ok && id.Name == "tests" {
						allowed = true
						break
					}
				}
				if allowed {
					break
				}
			}
			if allowed {
				break
			}
		}
	}
	if allowed {
		return v
	}

	v.pass.Report(analysis.Diagnostic{
		Pos:     cl.Lbrace,
		Message: "use field names when initializing structs; specify fields like `Field: value`",
	})
	return v
}

func (r *StructFieldKeyRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Walk(&structFieldVisitor{pass: pass}, file)
	}
	return nil, nil
}
