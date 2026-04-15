package rules

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

// LineLengthRule reports lines that exceed the soft recommended length.
type LineLengthRule struct{}

// BuildAnalyzer returns the analyzer for the line length rule
func (r *LineLengthRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "line_length",
		Doc: `detect lines that exceed the recommended 99 character soft limit.

This rule reports a diagnostic for any source line longer than 99 characters.
This is a stylistic suggestion (soft limit); if you disagree you can suppress
the check using ` + "//nolint:line_length" + ` on the offending line or file.
`,
		Run: r.run,
	}
}

func (r *LineLengthRule) run(pass *analysis.Pass) (any, error) {
	seen := make(map[string]bool)

	for _, f := range pass.Files {
		tf := pass.Fset.File(f.Pos())
		if tf == nil {
			continue
		}
		filename := tf.Name()
		if seen[filename] {
			continue
		}
		seen[filename] = true

		data, err := os.ReadFile(filename)
		if err != nil {
			// best-effort: skip files we cannot read
			continue
		}

		// Normalize CRLF and split into lines
		content := strings.ReplaceAll(string(data), "\r\n", "\n")
		lines := strings.Split(content, "\n")

		for i, line := range lines {
			if utf8.RuneCountInString(line) > 99 {
				// Line numbers are 1-based
				pos := tf.LineStart(i + 1)
				pass.Report(analysis.Diagnostic{
					Pos:     pos,
					Message: fmt.Sprintf("line exceeds recommended 99 character limit"),
				})
			}
		}
	}

	return nil, nil
}
