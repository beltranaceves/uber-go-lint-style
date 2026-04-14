package rules

import (
	"go/ast"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ImportGroupRule enforces two import groups: standard library first, then others.
type ImportGroupRule struct{}

func (r *ImportGroupRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "import_group",
		Doc: `Require two import groups: standard library first, then others.

This rule enforces that imports in a grouped import block are arranged with
standard library packages first (e.g. "fmt", "net/http"), then a single
blank line, then all third-party or internal imports (e.g. "go.uber.org/...", "github.com/...").`,
		Run: r.run,
	}
}

func isStdImport(pathLit string) bool {
	p := strings.Trim(pathLit, "\"")
	if p == "C" { // cgo pseudo-package
		return true
	}
	// Heuristic: standard library packages do not contain a dot in the path
	return !strings.Contains(p, ".")
}

func (r *ImportGroupRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.IMPORT {
				continue
			}

			// collect import specs in this declaration and sort by source position
			var specs []*ast.ImportSpec
			for _, s := range gen.Specs {
				if ispec, ok := s.(*ast.ImportSpec); ok {
					specs = append(specs, ispec)
				}
			}

			sort.Slice(specs, func(i, j int) bool {
				return specs[i].Pos() < specs[j].Pos()
			})

			if len(specs) <= 1 {
				// single import (or none) -- nothing to group
				continue
			}

			// Determine import ordering from the original source text
			fname := pass.Fset.Position(gen.Pos()).Filename
			var groups []bool
			var orderedPaths []string
			pathToSpec := make(map[string]int)
			data, err := os.ReadFile(fname)
			if err != nil {
				for _, s := range specs {
					groups = append(groups, isStdImport(s.Path.Value))
				}
			} else {
				start := pass.Fset.Position(gen.Pos()).Offset
				end := pass.Fset.Position(gen.End()).Offset
				if start < 0 || end > len(data) || start >= end {
					for _, s := range specs {
						groups = append(groups, isStdImport(s.Path.Value))
					}
				} else {
					snippet := string(data[start:end])
					re := regexp.MustCompile(`"([^"\\]+)"`)
					matches := re.FindAllStringSubmatch(snippet, -1)
					if len(matches) == 0 {
						for _, s := range specs {
							groups = append(groups, isStdImport(s.Path.Value))
						}
					} else {
						// Extract quoted import paths from each source line, skipping
						// comment-only lines. This avoids capturing quoted strings
						// from //want comments.
						lines := strings.Split(snippet, "\n")
						for _, ln := range lines {
							l := strings.TrimSpace(ln)
							if l == "" || strings.HasPrefix(l, "//") || l == "(" || l == ")" {
								continue
							}
							// find first quoted string in the line
							i := strings.Index(l, "\"")
							if i == -1 {
								continue
							}
							j := strings.Index(l[i+1:], "\"")
							if j == -1 {
								continue
							}
							p := l[i+1 : i+1+j]
							orderedPaths = append(orderedPaths, p)
						}
						// Build a map from path -> spec index for reporting
						for i, s := range specs {
							pathToSpec[strings.Trim(s.Path.Value, "\"")] = i
						}
						for _, p := range orderedPaths {
							groups = append(groups, isStdImport("\""+p+"\""))
						}
					}
				}
			}

			// Compress consecutive identical groups
			var comp []bool
			for i, g := range groups {
				if i == 0 || g != groups[i-1] {
					comp = append(comp, g)
				}
			}

			if len(comp) > 2 {
				// More than two groups -> report first spec of the third group
				// find index of first spec belonging to the third group
				targetGroup := comp[2]
				cur := comp[0]
				grpIndex := 0
				for i, g := range groups {
					if g != cur {
						grpIndex++
						cur = g
					}
					if grpIndex == 2 && g == targetGroup {
						if i < len(orderedPaths) {
							if sidx, ok := pathToSpec[orderedPaths[i]]; ok {
								pass.Report(analysis.Diagnostic{
									Pos:     specs[sidx].Pos(),
									Message: "imports must be grouped: standard library first, then third-party imports",
								})
							}
						}
						break
					}
				}
				continue
			}

			if len(comp) == 2 {
				// Expect first group to be std (true), second to be other (false)
				if !comp[0] {
					// first group is third-party -> report at first source-ordered spec
					if len(orderedPaths) > 0 {
						if sidx, ok := pathToSpec[orderedPaths[0]]; ok {
							pass.Report(analysis.Diagnostic{
								Pos:     specs[sidx].Pos(),
								Message: "imports must be grouped: standard library first, then third-party imports",
							})
						}
					}
					continue
				}

				// Check for blank line between groups using source-order indices
				lastStdSrc := -1
				firstOtherSrc := -1
				for i, g := range groups {
					if g {
						lastStdSrc = i
					} else if firstOtherSrc == -1 {
						firstOtherSrc = i
					}
				}
				if lastStdSrc != -1 && firstOtherSrc != -1 {
					// Map to spec indices
					if lastStdSrc < len(orderedPaths) && firstOtherSrc < len(orderedPaths) {
						lastPath := orderedPaths[lastStdSrc]
						firstPath := orderedPaths[firstOtherSrc]
						if lastIdx, ok1 := pathToSpec[lastPath]; ok1 {
							if firstIdx, ok2 := pathToSpec[firstPath]; ok2 {
								lastLine := pass.Fset.Position(specs[lastIdx].End()).Line
								firstLine := pass.Fset.Position(specs[firstIdx].Pos()).Line

								if firstLine-lastLine < 2 {
									pass.Report(analysis.Diagnostic{
										Pos:     specs[firstIdx].Pos(),
										Message: "add blank line between standard library and other imports",
									})
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, nil
}
