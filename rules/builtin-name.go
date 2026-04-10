package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// BuiltinNameRule checks that built-in names are not shadowed or used as identifiers.
type BuiltinNameRule struct{}

// builtinNames contains all predeclared identifiers that should not be shadowed.
// Computed from types.Universe to be automatically compatible with new Go versions.
var builtinNames = func() map[string]bool {
	names := make(map[string]bool)
	for _, name := range types.Universe.Names() {
		names[name] = true
	}
	return names
}()

// BuildAnalyzer returns the analyzer for the builtin-name rule
func (r *BuiltinNameRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "builtin_name",
		Doc: `avoid using predeclared identifiers for variable and field names.

			Go has several predeclared identifiers (types, constants, functions).
			Reusing these names as variable or field names can shadow the original within
			the current lexical scope and make code confusing or hard to grep.`,
		Run: r.run,
	}
}

func (r *BuiltinNameRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			// Check variable declarations: var x error
			case *ast.GenDecl:
				if node.Tok == token.VAR {
					for _, spec := range node.Specs {
						if valSpec, ok := spec.(*ast.ValueSpec); ok {
							for _, name := range valSpec.Names {
								if builtinNames[name.Name] {
									pass.Report(analysis.Diagnostic{
										Pos:     name.Pos(),
										Message: "identifier '" + name.Name + "' shadows a built-in, consider using a different name",
									})
								}
							}
						}
					}
				}

			// Check function parameters, receiver parameters, and struct fields
			case *ast.FuncDecl:
				// Check function parameters: func f(error string)
				if node.Type.Params != nil {
					for _, param := range node.Type.Params.List {
						for _, name := range param.Names {
							if builtinNames[name.Name] {
								pass.Report(analysis.Diagnostic{
									Pos:     name.Pos(),
									Message: "identifier '" + name.Name + "' shadows a built-in, consider using a different name",
								})
							}
						}
					}
				}

				// Check receiver parameters: func (error string) Method()
				if node.Recv != nil {
					for _, param := range node.Recv.List {
						for _, name := range param.Names {
							if builtinNames[name.Name] {
								pass.Report(analysis.Diagnostic{
									Pos:     name.Pos(),
									Message: "identifier '" + name.Name + "' shadows a built-in, consider using a different name",
								})
							}
						}
					}
				}

			// Check struct fields: type Foo struct { error error }
			case *ast.StructType:
				if node.Fields != nil {
					for _, field := range node.Fields.List {
						for _, name := range field.Names {
							if builtinNames[name.Name] {
								pass.Report(analysis.Diagnostic{
									Pos:     name.Pos(),
									Message: "identifier '" + name.Name + "' shadows a built-in, consider using a different name",
								})
							}
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}
