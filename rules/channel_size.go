package rules

import (
	"go/ast"
	"strconv"
	"strings"

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
		ast.Inspect(file, func(n ast.Node) bool {
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
					pass.Report(analysis.Diagnostic{
						Pos:     capExpr.Pos(),
						Message: "channel size should be one or unbuffered",
					})
					return true
				}

				if i, err := strconv.Atoi(v); err == nil {
					if i == 0 || i == 1 {
						return true
					}
					pass.Report(analysis.Diagnostic{
						Pos:     capExpr.Pos(),
						Message: "channel size should be one or unbuffered",
					})
					return true
				}

				// couldn't parse literal -> report conservatively
				pass.Report(analysis.Diagnostic{
					Pos:     capExpr.Pos(),
					Message: "channel size should be one or unbuffered",
				})
				return true
			}

			// Non-literal capacity (variable or expression) — flag for scrutiny
			pass.Report(analysis.Diagnostic{
				Pos:     capExpr.Pos(),
				Message: "channel size should be one or unbuffered",
			})

			return true
		})
	}

	return nil, nil
}
