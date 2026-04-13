package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// EnumStartRule enforces that enum-like const groups start at 1 instead of 0.
type EnumStartRule struct{}

// BuildAnalyzer returns the analyzer for enum_start rule
func (r *EnumStartRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "enum_start",
		Doc: `start enums at one.

This rule flags const groups that define a named integer type using ` + "`iota`" + ` where the first enumerator value is 0. Prefer starting enums at 1 so the zero value remains an invalid/default sentinel.
`,
		Run: r.run,
	}
}

func exprUsesIotaWithValue(expr ast.Expr) (usesIota bool, valueIsZero bool, ok bool) {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == "iota" {
			return true, true, true
		}
	case *ast.BasicLit:
		// numeric literal
		if e.Kind == token.INT && e.Value == "0" {
			return false, true, true
		}
	case *ast.BinaryExpr:
		// look for iota + N
		if id, okid := e.X.(*ast.Ident); okid && id.Name == "iota" {
			if lit, oklit := e.Y.(*ast.BasicLit); oklit && lit.Kind == token.INT {
				if lit.Value == "0" {
					return true, true, true
				}
				return true, false, true
			}
		}
		if id, okid := e.Y.(*ast.Ident); okid && id.Name == "iota" {
			if lit, oklit := e.X.(*ast.BasicLit); oklit && lit.Kind == token.INT {
				if lit.Value == "0" {
					return true, true, true
				}
				return true, false, true
			}
		}
	}
	return false, false, false
}

func (r *EnumStartRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST {
				return true
			}

			// iterate specs in order to find the first enumerator that has a named integer type
			for i, s := range decl.Specs {
				vs, ok := s.(*ast.ValueSpec)
				if !ok || len(vs.Names) == 0 {
					continue
				}

				// Use types info to determine the declared type of the constant name
				name := vs.Names[0]
				obj := pass.TypesInfo.Defs[name]
				if obj == nil {
					continue
				}

				typ := obj.Type()
				named, ok := typ.(*types.Named)
				if !ok {
					continue
				}

				// only consider named integer-like underlying types
				switch named.Underlying().(type) {
				case *types.Basic:
					// allowed
				default:
					continue
				}

				// We only analyze the first spec that defines this enum in the group
				if i >= len(decl.Specs) {
					// defensive
				}

				// Determine start expression if present
				if len(vs.Values) == 0 {
					// no explicit value on this spec: skip (can't deterministically compute start)
					break
				}

				expr := vs.Values[0]
				usesIota, isZero, okv := exprUsesIotaWithValue(expr)
				if okv {
					if (usesIota && isZero) || (!usesIota && isZero) {
						pass.Report(analysis.Diagnostic{
							Pos:     vs.Pos(),
							Message: "enum '" + named.Obj().Name() + "' starts at 0; prefer starting at 1 so the zero value is not a valid member (use 'iota + 1' or add an explicit sentinel).",
						})
					}
				}

				// only examine the first matched spec for this group/type
				break
			}

			return true
		})
	}

	return nil, nil
}
