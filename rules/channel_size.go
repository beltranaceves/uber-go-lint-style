package rules

import (
	"go/ast"
	"strconv"
	"strings"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// ChannelSizeRule enforces that channels are either unbuffered or have size one.
type ChannelSizeRule struct{}

// BuildAnalyzer returns the analyzer for the channel_size rule
func (r *ChannelSizeRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "channel_size",
		Doc: `channels should usually be unbuffered or sized to one.

This rule reports uses of make(chan T, N) where N is not 0 or 1. Any other
buffer sizes should be subject to careful review and documented justification.`,
		Run: r.run,
	}
}

func (r *ChannelSizeRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// Inspect function declarations individually so we can skip bodies
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if hasNolintCommentGroup(fn.Doc, "channel_size") {
					// skip this function body entirely
					continue
				}

				if fn.Body == nil {
					continue
				}

				ast.Inspect(fn.Body, func(n ast.Node) bool {
					callExpr, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}

					// Look for builtin make(...) calls
					ident, ok := callExpr.Fun.(*ast.Ident)
					if !ok || ident.Name != "make" {
						return true
					}

					// make takes a type as the first arg; for channels this is a *ast.ChanType
					if len(callExpr.Args) < 2 {
						return true
					}

					// Only interested in channels: first arg should be a ChanType
					if _, ok := callExpr.Args[0].(*ast.ChanType); !ok {
						return true
					}

					// Second arg is capacity; if missing it's unbuffered. If present, check value.
					capExpr := callExpr.Args[1]

					// If it's a basic literal, try to parse the integer
					if bl, ok := capExpr.(*ast.BasicLit); ok {
						// strip underscores allowed in numeric literals
						v := strings.ReplaceAll(bl.Value, "_", "")
						// basic lit for ints may be decimal, hex, etc. Try Atoi for decimal only.
						if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
							// hex -> treat as non-allowed size (report)
							if !isNolint(pass, file, capExpr.Pos(), "channel_size") {
								pass.Report(analysis.Diagnostic{
									Pos:     capExpr.Pos(),
									Message: "channel size should be one or unbuffered",
								})
							}
							return true
						}

						if i, err := strconv.Atoi(v); err == nil {
							if i == 0 || i == 1 {
								return true
							}
							if !isNolint(pass, file, capExpr.Pos(), "channel_size") {
								pass.Report(analysis.Diagnostic{
									Pos:     capExpr.Pos(),
									Message: "channel size should be one or unbuffered",
								})
							}
							return true
						}

						// couldn't parse literal -> report conservatively
						if !isNolint(pass, file, capExpr.Pos(), "channel_size") {
							pass.Report(analysis.Diagnostic{
								Pos:     capExpr.Pos(),
								Message: "channel size should be one or unbuffered",
							})
						}
						return true
					}

					// Non-literal capacity (variable or expression) — flag for scrutiny
					if !isNolint(pass, file, capExpr.Pos(), "channel_size") {
						pass.Report(analysis.Diagnostic{
							Pos:     capExpr.Pos(),
							Message: "channel size should be one or unbuffered",
						})
					}

					return true
				})

			} else {
				// Non-function declarations: inspect for make calls (package-level inits)
				ast.Inspect(decl, func(n ast.Node) bool {
					callExpr, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}

					ident, ok := callExpr.Fun.(*ast.Ident)
					if !ok || ident.Name != "make" {
						return true
					}

					if len(callExpr.Args) < 2 {
						return true
					}

					if _, ok := callExpr.Args[0].(*ast.ChanType); !ok {
						return true
					}

					capExpr := callExpr.Args[1]
					if !isNolint(pass, file, capExpr.Pos(), "channel_size") {
						pass.Report(analysis.Diagnostic{
							Pos:     capExpr.Pos(),
							Message: "channel size should be one or unbuffered",
						})
					}

					return true
				})
			}
		}
	}

	return nil, nil
}

// isNolint checks whether a comment on the same line or the previous line
// contains a nolint directive for the given rule name or a generic nolint.
func isNolint(pass *analysis.Pass, file *ast.File, pos token.Pos, rule string) bool {
	p := pass.Fset.Position(pos)
	// iterate comment groups in the file
	for _, cg := range file.Comments {
		if cg == nil || len(cg.List) == 0 {
			continue
		}
		start := pass.Fset.Position(cg.Pos()).Line
		end := pass.Fset.Position(cg.End()).Line
		// consider comments on the same line or up to two lines above
		if p.Line < start-2 || p.Line > end {
			continue
		}
		text := ""
		for _, c := range cg.List {
			text += c.Text + "\n"
		}
		// Normalize and check for nolint
		lower := strings.ToLower(text)
		if strings.Contains(lower, "nolint") {
			// If specific rule present: nolint: channel_size
			if strings.Contains(lower, "nolint:") {
				// split after nolint:
				parts := strings.SplitN(lower, "nolint:", 2)
				if len(parts) == 2 {
					// check if rule name is listed
					if strings.Contains(parts[1], rule) {
						return true
					}
				}
			}
			// generic nolint
			if strings.Contains(lower, "// nolint") || strings.Contains(lower, "/* nolint") {
				return true
			}
		}
	}

	return false
}

func hasNolintCommentGroup(cg *ast.CommentGroup, rule string) bool {
	if cg == nil {
		return false
	}
	text := ""
	for _, c := range cg.List {
		text += c.Text + "\n"
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "nolint") {
		if strings.Contains(lower, "nolint:") {
			parts := strings.SplitN(lower, "nolint:", 2)
			if len(parts) == 2 {
				if strings.Contains(parts[1], rule) {
					return true
				}
			}
		}
		if strings.Contains(lower, "// nolint") || strings.Contains(lower, "/* nolint") {
			return true
		}
	}
	return false
}
