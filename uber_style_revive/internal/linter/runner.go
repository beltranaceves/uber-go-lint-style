package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beltranaceves/uber-go-lint-style/uber_style_revive/rules"
	"github.com/mgechev/revive/lint"
)

// Runner orchestrates linting with Uber rules
type Runner struct {
	config *lint.Config
	linter *lint.Linter
}

// NewRunner creates a new linter runner
func NewRunner() *Runner {
	linter := lint.New(os.ReadFile, 0)

	config := &lint.Config{
		Confidence: 0,
		Rules:      make(lint.RulesConfig),
	}

	// Configure all rules
	for _, name := range rules.GetAllRuleNames() {
		config.Rules[name] = lint.RuleConfig{
			Arguments: nil,
		}
	}

	return &Runner{
		config: config,
		linter: linter,
	}
}

// LintPaths lints the provided file paths
func (r *Runner) LintPaths(paths []string) ([]lint.Failure, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no paths provided")
	}

	// Convert individual paths to file pattern format expected by revive
	var filePatterns [][]string
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			// Directory: add all .go files
			filePatterns = append(filePatterns, []string{filepath.Join(path, "...")})
		} else {
			// File or pattern
			filePatterns = append(filePatterns, []string{path})
		}
	}

	// Run linter with all Uber rules
	failuresCh, err := r.linter.Lint(filePatterns, rules.GetAllRules(), *r.config)
	if err != nil {
		return nil, fmt.Errorf("linting failed: %w", err)
	}

	var failures []lint.Failure
	for failure := range failuresCh {
		failures = append(failures, failure)
	}

	return failures, nil
}

// LintFile lints a single file
func (r *Runner) LintFile(filePath string) ([]lint.Failure, error) {
	return r.LintPaths([]string{filePath})
}

// FormatFailures formats lint failures for output
func FormatFailures(failures []lint.Failure, format string) string {
	if len(failures) == 0 {
		return ""
	}

	switch format {
	case "json":
		return formatJSON(failures)
	case "simple":
		return formatSimple(failures)
	default:
		return formatFriendly(failures)
	}
}

func formatFriendly(failures []lint.Failure) string {
	var output strings.Builder

	for _, failure := range failures {
		output.WriteString(fmt.Sprintf("%s:%d:%d: %s [%s]\n",
			failure.GetFilename(),
			failure.GetLine(),
			failure.GetColumn(),
			failure.Failure,
			failure.RuleName,
		))
	}

	return output.String()
}

func formatSimple(failures []lint.Failure) string {
	var output strings.Builder

	for _, failure := range failures {
		output.WriteString(fmt.Sprintf("%s:%d %s\n",
			failure.GetFilename(),
			failure.GetLine(),
			failure.Failure,
		))
	}

	return output.String()
}

func formatJSON(failures []lint.Failure) string {
	var output strings.Builder
	output.WriteString("[\n")

	for i, failure := range failures {
		output.WriteString(fmt.Sprintf("  {\n"))
		output.WriteString(fmt.Sprintf("    \"file\": \"%s\",\n", failure.GetFilename()))
		output.WriteString(fmt.Sprintf("    \"line\": %d,\n", failure.GetLine()))
		output.WriteString(fmt.Sprintf("    \"column\": %d,\n", failure.GetColumn()))
		output.WriteString(fmt.Sprintf("    \"rule\": \"%s\",\n", failure.RuleName))
		output.WriteString(fmt.Sprintf("    \"message\": \"%s\"\n", failure.Failure))
		output.WriteString("  }")

		if i < len(failures)-1 {
			output.WriteString(",")
		}
		output.WriteString("\n")
	}

	output.WriteString("]\n")
	return output.String()
}
