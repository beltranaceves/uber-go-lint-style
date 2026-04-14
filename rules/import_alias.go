package rules

import (
	"fmt"
	"path"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ImportAliasRule enforces the project's import aliasing policy.
type ImportAliasRule struct{}

func (r *ImportAliasRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "import_alias",
		Doc: `Import aliasing must be used if the package name does not match
the last element of the import path. In all other scenarios, import
aliases should be avoided unless there is a direct conflict between
imports.`,
		Run: r.run,
	}
}

func (r *ImportAliasRule) run(pass *analysis.Pass) (any, error) {
	// Map import path -> declared package name (from types information)
	pkgByPath := make(map[string]string)
	for _, ipkg := range pass.Pkg.Imports() {
		pkgByPath[ipkg.Path()] = ipkg.Name()
	}

	// Count how many imports share the same declared package name
	defaultNameCount := make(map[string]int)
	for _, file := range pass.Files {
		for _, ispec := range file.Imports {
			if ispec.Name != nil {
				// Skip blank and dot imports for counting
				if ispec.Name.Name == "_" || ispec.Name.Name == "." {
					continue
				}
			}
			importPath := strings.Trim(ispec.Path.Value, "\"")
			declared := ""
			if n, ok := pkgByPath[importPath]; ok {
				declared = n
			} else {
				declared = path.Base(importPath)
			}
			defaultNameCount[declared]++
		}
	}

	for _, file := range pass.Files {
		for _, ispec := range file.Imports {
			// Ignore blank and dot imports
			if ispec.Name != nil {
				if ispec.Name.Name == "_" || ispec.Name.Name == "." {
					continue
				}
			}

			importPath := strings.Trim(ispec.Path.Value, "\"")
			last := path.Base(importPath)
			declared := ""
			if n, ok := pkgByPath[importPath]; ok {
				declared = n
			} else {
				declared = last
			}

			if ispec.Name == nil {
				// No alias provided: require alias when declared package name != last path element
				if declared != last {
					pass.Report(analysis.Diagnostic{
						Pos:     ispec.Pos(),
						Message: fmt.Sprintf("import path \"%s\" package name \"%s\" does not match last path element \"%s\"; add an explicit alias \"%s\"", importPath, declared, last, declared),
					})
				}
			} else {
				alias := ispec.Name.Name
				// If declared==last and there is no conflict, alias is unnecessary
				if declared == last {
					if defaultNameCount[declared] <= 1 {
						pass.Report(analysis.Diagnostic{
							Pos:     ispec.Name.Pos(),
							Message: fmt.Sprintf("unnecessary import alias '%s' for package '%s'; remove the alias", alias, declared),
						})
					}
				}
			}
		}
	}

	return nil, nil
}
