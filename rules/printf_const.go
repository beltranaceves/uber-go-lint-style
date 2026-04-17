package rules

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// PrintfConstRule enforces that format strings used in Printf-style calls are const.
type PrintfConstRule struct{}

// BuildAnalyzer returns the analyzer for the printf_const rule
func (r *PrintfConstRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "printf_const",
		Doc:  "ensure format strings passed to fmt.Printf-style functions are declared as const",
		Run:  r.run,
	}
}

func (r *PrintfConstRule) run(pass *analysis.Pass) (any, error) {
	// Map function name -> index of the format argument
	idxMap := map[string]int{
		"Printf":  0,
		"Sprintf": 0,
		"Errorf":  0,
		"Fprintf": 1,
	}

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

			name := sel.Sel.Name
			idx, ok := idxMap[name]
			if !ok {
				return true
			}

			// Ensure the selector resolves to a function in package fmt.
			if obj := pass.TypesInfo.Uses[sel.Sel]; obj != nil {
				if fn, ok := obj.(*types.Func); ok {
					if fn.Pkg() == nil || fn.Pkg().Path() != "fmt" {
						return true
					}

					if idx >= len(call.Args) {
						return true
					}

					switch a := call.Args[idx].(type) {
					case *ast.BasicLit:
						// literal format string - fine
					case *ast.Ident:
						if usedObj := pass.TypesInfo.Uses[a]; usedObj != nil {
							if _, isConst := usedObj.(*types.Const); !isConst {
								pass.Report(analysis.Diagnostic{
									Pos:      a.Pos(),
									End:      a.End(),
									Category: "printf_const",
									Message:  "format string should be a const value",
								})
							}
						}
					default:
						// other expressions (calls, selectors, etc.) - do not warn
					}
				}
			}

			return true
		})
	}

	return nil, nil
}
