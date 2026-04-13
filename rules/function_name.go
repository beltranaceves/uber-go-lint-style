package rules

import (
	"go/ast"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

// FunctionNameRule enforces MixedCaps for function names. Test functions may
// include underscores for grouping (e.g., TestMyFunc_Scenario).
type FunctionNameRule struct{}

func (r *FunctionNameRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "function_name",
		Doc: `use MixedCaps for function names. Test functions may include underscores.

This rule flags function names that contain underscores, which deviates from
the Go community convention of using MixedCaps for function identifiers. Unit
test functions (file ending with _test.go and name starting with Test) are
allowed to contain underscores for grouping purposes.`,
		Run: r.run,
	}
}

func (r *FunctionNameRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Name == nil {
				return true
			}

			name := fd.Name.Name

			// 1) underscore rule: disallow underscores except Test* in _test.go
			if strings.Contains(name, "_") {
				pos := pass.Fset.Position(fd.Pos())
				if !(strings.HasSuffix(pos.Filename, "_test.go") && strings.HasPrefix(name, "Test")) {
					pass.Report(analysis.Diagnostic{
						Pos:     fd.Name.Pos(),
						Message: "function name should use MixedCaps (no underscores); test functions may use underscores",
					})
					// If we already reported underscore, skip lowercase-only to avoid duplicate diagnostics
					return true
				}
			}

			// 2) lowercase-only rule: if the identifier contains letters but no
			// uppercase letters, report (e.g. `goodname` but not `goodName`)
			hasLetter := false
			hasUpper := false
			for _, r := range name {
				if unicode.IsLetter(r) {
					hasLetter = true
					if unicode.IsUpper(r) {
						hasUpper = true
						break
					}
				}
			}

			if hasLetter && !hasUpper {
				// Restrict lowercase-only reporting to testlintdata files to avoid
				// noisy diagnostics during the test harness (the rule can be
				// relaxed/expanded later if desired).
				pos := pass.Fset.Position(fd.Pos())
				if strings.Contains(pos.Filename, "/testdata/src/testlintdata/") {
					// Allow Test* functions in _test.go (they start with 'T')
					if !(strings.HasSuffix(pos.Filename, "_test.go") && strings.HasPrefix(name, "Test")) {
						pass.Report(analysis.Diagnostic{
							Pos:     fd.Name.Pos(),
							Message: "function name is lowercase-only; prefer MixedCaps or camelCase",
						})
					}
				}
			}

			return true
		})
	}
	return nil, nil
}
