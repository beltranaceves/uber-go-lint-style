package rules

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// StringEscapeRule suggests using raw string literals when a double-quoted
// literal contains hand-escaped quotes and doesn't rely on escape sequences.
type StringEscapeRule struct{}

func (r *StringEscapeRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "string_escape",
		Doc: `use raw string literals to avoid hand-escaped quotes.

This rule detects double-quoted string literals that embed escaped double
quotes (\") and that do not rely on other escape sequences (for example
"\n"), and recommends using a raw string literal (backticks) instead.

Limitations: raw string literals cannot contain backticks and do not honor
escape sequences. The rule intentionally skips strings that contain
escape sequences such as \n, \t, or backslashes, or strings whose unquoted
content contains a backtick.`,
		Run: r.run,
	}
}

func (r *StringEscapeRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			bl, ok := n.(*ast.BasicLit)
			if !ok || bl.Kind != token.STRING {
				return true
			}

			src := bl.Value // quoted literal including surrounding quotes

			// Only consider interpreted string literals (start and end with ")
			if len(src) < 2 || src[0] != '"' {
				return true
			}

			// Must contain an escaped quote (\") in source to consider.
			if !strings.Contains(src, `\"`) && !strings.Contains(src, `"`) {
				return true
			}

			// If the literal contains common escape sequences other than escaped quote,
			// skip (raw literal would change semantics).
			escapeSeqs := []string{"\\n", "\\r", "\\t", "\\b", "\\f", "\\v", "\\x", "\\u", "\\U", "\\\\"}
			for _, e := range escapeSeqs {
				if strings.Contains(src, e) {
					return true
				}
			}

			// Unquote to inspect content (and to check for backticks which raw
			// literals cannot contain).
			unq, err := strconv.Unquote(src)
			if err != nil {
				return true
			}
			if strings.Contains(unq, "`") {
				return true
			}

			pass.Report(analysis.Diagnostic{
				Pos:     bl.Pos(),
				Message: "use raw string literal to avoid escaping quotes",
			})

			return true
		})
	}
	return nil, nil
}
