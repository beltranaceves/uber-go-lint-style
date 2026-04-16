package rules

import (
	"fmt"
	"go/ast"
	"go/token"

	"go/types"

	"golang.org/x/tools/go/analysis"
)

// ParamNakedRule detects naked boolean parameters passed as plain literals
// at call sites (e.g., foo(true)). It encourages using a comment like
// /* paramName */ or a named type for clarity.
type ParamNakedRule struct{}

func (r *ParamNakedRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "param_naked",
		Doc: `avoid naked parameters in function calls.

This rule reports boolean literal arguments (true/false) passed to
function parameters. Prefer adding an inline comment (/* paramName */)
or using a named type for the parameter to improve call-site readability.
`,
		Run: r.run,
	}
}

func (r *ParamNakedRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// capture file-level comments for inline-comment checks
		commentGroups := file.Comments

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			var sig *types.Signature

			switch fun := call.Fun.(type) {
			case *ast.SelectorExpr:
				if obj := pass.TypesInfo.Uses[fun.Sel]; obj != nil {
					if fn, ok := obj.(*types.Func); ok {
						if s, ok := fn.Type().(*types.Signature); ok {
							sig = s
						}
					}
				}
			case *ast.Ident:
				if obj := pass.TypesInfo.ObjectOf(fun); obj != nil {
					if fn, ok := obj.(*types.Func); ok {
						if s, ok := fn.Type().(*types.Signature); ok {
							sig = s
						}
					}
				}
			}

			// If we couldn't get a signature from the object, try the type of the function expression.
			if sig == nil {
				if t := pass.TypesInfo.TypeOf(call.Fun); t != nil {
					if s, ok := t.(*types.Signature); ok {
						sig = s
					}
				}
			}

			if sig == nil {
				return true
			}

			params := sig.Params()
			isVariadic := sig.Variadic()

			for i, arg := range call.Args {
				paramIdx := i
				if isVariadic && i >= params.Len()-1 {
					paramIdx = params.Len() - 1
				}
				if paramIdx < 0 || paramIdx >= params.Len() {
					continue
				}

				p := params.At(paramIdx)
				// Check whether the parameter type is bool (or underlying bool)
				if isBoolType(p.Type()) {
					// Check whether the argument is a boolean literal (true/false)
					if ident, ok := arg.(*ast.Ident); ok {
						if ident.Name == "true" || ident.Name == "false" {
							// Skip if there's an inline trailing comment annotating the arg
							hasInline := false
							endPos := call.End()
							if call.Rparen.IsValid() {
								endPos = call.Rparen
							}
							for _, cg := range commentGroups {
								if cg.Pos() > arg.End() && cg.Pos() < endPos {
									// only consider same-line inline comments
									if pass.Fset.Position(cg.Pos()).Line == pass.Fset.Position(arg.End()).Line {
										hasInline = true
										break
									}
								}
							}
							if hasInline {
								continue
							}

							paramName := p.Name()
							if paramName == "" {
								paramName = "parameter"
							}
							msg := fmt.Sprintf("avoid naked boolean parameter %q; add an inline comment at callsite or use a named type", paramName)
							pass.Report(analysis.Diagnostic{
								Pos:     arg.Pos(),
								End:     token.Pos(arg.End()),
								Message: msg,
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

func isBoolType(t types.Type) bool {
	if t == nil {
		return false
	}
	if b, ok := t.Underlying().(*types.Basic); ok {
		return b.Kind() == types.Bool
	}
	return false
}
