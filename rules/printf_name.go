package rules

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// PrintfNameRule enforces that printf-style functions end with 'f'.
type PrintfNameRule struct{}

// BuildAnalyzer returns the analyzer for the printf_name rule
func (r *PrintfNameRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "printf_name",
		Doc: `printf-style functions should be named with a trailing 'f'.

This rule detects functions that accept a format string and a variadic
parameter (e.g. ` + "`format string, a ...interface{}`" + `) and reports if the
function name does not end with 'f'. This helps ` + "`go vet`" + ` and other tools
recognize printf-style functions and makes the naming consistent.
`,
		Run: r.run,
	}
}

func (r *PrintfNameRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Type == nil || fd.Type.Params == nil {
				return true
			}

			params := fd.Type.Params.List
			if len(params) == 0 {
				return true
			}

			// require at least one string parameter anywhere
			hasString := false
			for _, p := range params {
				if ident, ok := p.Type.(*ast.Ident); ok && ident.Name == "string" {
					hasString = true
					break
				}
			}
			if !hasString {
				return true
			}

			// require last parameter to be variadic and of interface{} / any
			last := params[len(params)-1]
			ell, ok := last.Type.(*ast.Ellipsis)
			if !ok {
				return true
			}

			isVariadicInterface := false
			switch t := ell.Elt.(type) {
			case *ast.InterfaceType:
				isVariadicInterface = true
			case *ast.Ident:
				// accept `any` as well
				if t.Name == "any" {
					isVariadicInterface = true
				}
			}
			if !isVariadicInterface {
				return true
			}

			name := fd.Name.Name
			if strings.HasSuffix(name, "f") {
				return true
			}

			suggested := name + "f"
			pass.Report(analysis.Diagnostic{
				Pos:     fd.Name.Pos(),
				Message: "printf-style function '" + name + "' should be named '" + suggested + "'",
			})

			return true
		})
	}
	return nil, nil
}
