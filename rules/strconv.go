package rules

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// StrconvRule flags uses of fmt.Sprint/Sprintln/Sprintf that convert
// primitive types to strings and recommends using strconv instead.
type StrconvRule struct{}

func (r *StrconvRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "prefer_strconv",
		Doc: `prefer strconv over fmt for primitive-to-string conversions.

This rule flags uses of fmt.Sprint, fmt.Sprintln, and fmt.Sprintf when they are
used to convert primitive types (bool, integers, unsigned integers, floats)
to strings and recommends using functions from the strconv package which are
more efficient for these conversions. The analyzer requires type information.
`,
		Run: r.run,
	}
}

func (r *StrconvRule) run(pass *analysis.Pass) (any, error) {
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
			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok || pkgIdent.Name != "fmt" {
				return true
			}
			name := sel.Sel.Name
			if name != "Sprint" && name != "Sprintln" && name != "Sprintf" {
				return true
			}

			// Sprint/Sprintln: consider only single-argument conversions
			if (name == "Sprint" || name == "Sprintln") && len(call.Args) != 1 {
				return true
			}

			// Sprintf: require a string literal format and exactly one value arg
			if name == "Sprintf" {
				if len(call.Args) != 2 {
					return true
				}
				if _, ok := call.Args[0].(*ast.BasicLit); !ok {
					return true
				}
			}

			var arg ast.Expr
			if name == "Sprintf" {
				arg = call.Args[1]
			} else {
				arg = call.Args[0]
			}

			t := pass.TypesInfo.TypeOf(arg)
			if t == nil {
				return true
			}
			if isPrimitive(t) {
				pass.Report(analysis.Diagnostic{
					Pos:     call.Pos(),
					Message: "prefer strconv functions for primitive-to-string conversions instead of fmt." + name,
				})
			}

			return true
		})
	}
	return nil, nil
}

func isPrimitive(t types.Type) bool {
	if t == nil {
		return false
	}
	if basic, ok := t.Underlying().(*types.Basic); ok {
		switch basic.Kind() {
		case types.Bool,
			types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
			types.Float32, types.Float64:
			return true
		}
	}
	return false
}
