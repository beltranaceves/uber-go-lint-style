package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	rules := []string{
		"interface-pointer", "interface-compliance", "interface-receiver",
		"mutex-zero-value", "container-copy", "defer-clean", "channel-size",
		"enum-start", "time", "type-assert", "panic", "atomic", "global-mut",
		"embed-public", "builtin-name", "init", "exit-main", "exit-once",
		"struct-tag", "goroutine-forget", "goroutine-exit", "goroutine-init",
		"strconv", "string-byte-slice", "container-capacity",
		"line-length", "consistency", "decl-group", "import-group",
		"package-name", "function-name", "import-alias", "function-order",
		"nest-less", "else-unnecessary", "global-decl", "global-name",
		"struct-embed", "var-decl", "slice-nil", "var-scope", "param-naked",
		"string-escape", "struct-field-key", "struct-field-zero", "struct-zero",
		"struct-pointer", "map-init", "printf-const", "printf-name",
		"test-table", "functional-option", "lint",
		"error-type", "error-wrap", "error-name", "error-once",
		"performance",
	}

	rulesDir := "rules"
	os.MkdirAll(rulesDir, 0755)

	for _, ruleName := range rules {
		// Convert rule-name to RuleName (e.g., interface-pointer -> InterfacePointer)
		ruleNameCamel := toCamelCase(ruleName)

		filename := filepath.Join(rulesDir, ruleName+".go")

		// Check if file already exists
		if _, err := os.Stat(filename); err == nil {
			fmt.Printf("Skipping %s (already exists)\n", ruleName)
			continue
		}

		content := fmt.Sprintf(`package rules

import (
	"github.com/mgechev/revive/lint"
)

// %[1]sRule lints %s according to Uber Go Style Guide.
// Reference: https://github.com/uber-go/guide/blob/master/style.md#%s
type %[1]sRule struct{}

// Name returns the rule name
func (r *%[1]sRule) Name() string {
	return "%s"
}

// Apply runs the rule against the provided file
func (r *%[1]sRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	// TODO: Implement rule logic
	// - Read file.AST to traverse the Go code
	// - Find patterns that violate the style guide
	// - Return lint.Failure for each violation
	return failures
}
`, ruleNameCamel, ruleName, ruleName, ruleName)

		os.WriteFile(filename, []byte(content), 0644)
		fmt.Printf("Created %s\n", ruleName)
	}

	fmt.Printf("\nTotal: %d rules created\n", len(rules))
}

func toCamelCase(s string) string {
	// Split on hyphens
	parts := strings.Split(s, "-")

	// Capitalize first letter of each part
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, "")
}
