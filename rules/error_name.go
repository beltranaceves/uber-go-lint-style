package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// ErrorNameRule enforces naming conventions for error variables and types.
type ErrorNameRule struct{}

func (r *ErrorNameRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "error_name",
		Doc: `enforce error naming: global error variables should be prefixed
with Err (exported) or err (unexported). Custom error types should be
named with the suffix Error.

This rule uses types information to detect variables of type error and
to identify types that implement the Error() string method.`,
		Run:      r.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func (r *ErrorNameRule) run(pass *analysis.Pass) (any, error) {
	// Iterate top-level declarations to find package-level vars and types.
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			// Package-level var declarations
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
				for _, spec := range gen.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							obj := pass.TypesInfo.Defs[name]
							if obj == nil {
								continue
							}
							// Check if the variable has type 'error'
							if obj.Type() != nil && obj.Type().String() == "error" {
								ident := name.Name
								if ast.IsExported(ident) {
									if !strings.HasPrefix(ident, "Err") {
										pass.Report(analysis.Diagnostic{
											Pos:     name.Pos(),
											Message: "exported error variable should be prefixed with 'Err'",
										})
									}
								} else {
									if !strings.HasPrefix(ident, "err") {
										pass.Report(analysis.Diagnostic{
											Pos:     name.Pos(),
											Message: "unexported error variable should be prefixed with 'err'",
										})
									}
								}
							}
						}
					}
				}
			}

			// Type declarations: check for types that implement Error() string
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
				for _, spec := range gen.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						name := ts.Name.Name
						// Look up the named type object in TypesInfo
						obj := pass.TypesInfo.Defs[ts.Name]
						if obj == nil {
							continue
						}
						named, ok := obj.Type().(*types.Named)
						if !ok {
							continue
						}
						// Check if the type has an Error method
						hasError := false
						for i := 0; i < named.NumMethods(); i++ {
							m := named.Method(i)
							if m.Name() == "Error" {
								// Check signature: first result string
								sig, ok := m.Type().(*types.Signature)
								if ok && sig.Results().Len() >= 1 {
									if sig.Results().At(0).Type().String() == "string" {
										hasError = true
										break
									}
								}
							}
						}
						if hasError {
							if !strings.HasSuffix(name, "Error") {
								pass.Report(analysis.Diagnostic{
									Pos:     ts.Name.Pos(),
									Message: "error type names should end with 'Error'",
								})
							}
						}
					}
				}
			}
		}
	}

	return nil, nil
}
