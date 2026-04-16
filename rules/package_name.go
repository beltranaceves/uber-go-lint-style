package rules

import (
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

// PackageNameRule enforces package naming conventions from the style guide.
type PackageNameRule struct{}

// BuildAnalyzer returns the analyzer for package-name rule.
func (r *PackageNameRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "package_name",
		Doc: `enforce package naming conventions: lower-case, no underscores, not generic or plural.

This rule checks the package name and reports when it:
- contains upper-case letters or underscores
- is one of discouraged generic names (common, util, shared, lib)
- appears to be plural (naive check: ends with 's')

The rule is opinionated; use //nolint:package_name to suppress where inappropriate.`,
		Run: r.run,
	}
}

func (r *PackageNameRule) run(pass *analysis.Pass) (any, error) {
	if pass.Pkg == nil {
		return nil, nil
	}

	name := pass.Pkg.Name()
	if name == "" {
		return nil, nil
	}

	// choose a position to report: first file's package identifier
	var pos token.Pos
	if len(pass.Files) > 0 && pass.Files[0].Name != nil {
		pos = pass.Files[0].Name.Pos()
	}

	// Check uppercase or underscore
	for _, rch := range name {
		if unicode.IsUpper(rch) || rch == '_' {
			pass.Report(analysis.Diagnostic{
				Pos:     pos,
				Message: "package name '" + name + "' should be lower-case and contain no underscores",
			})
			return nil, nil
		}
	}

	// Discouraged generic names
	switch name {
	case "common", "util", "shared", "lib":
		pass.Report(analysis.Diagnostic{
			Pos:     pos,
			Message: "package name '" + name + "' is discouraged; choose a more specific name",
		})
		return nil, nil
	}

	// Naive plural check
	if strings.HasSuffix(name, "s") && len(name) > 1 {
		pass.Report(analysis.Diagnostic{
			Pos:     pos,
			Message: "package name '" + name + "' should not be plural; prefer the singular form",
		})
		return nil, nil
	}

	return nil, nil
}
