package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/beltranaceves/uber-go-lint-style/internal/agents"
)

func main() {
	// List of all rules
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

	statusDir := ".agent-status"

	// Remove old status if exists
	os.RemoveAll(statusDir)

	af := agents.NewAgentFramework(statusDir)
	af.InitializeAgents(rules)

	fmt.Printf("Initialized %d agents\n", len(rules))
	fmt.Printf("Status directory: %s\n", statusDir)

	// List the status files
	entries, _ := os.ReadDir(statusDir)
	fmt.Printf("\nStatus files created:\n")
	for _, entry := range entries {
		fmt.Printf("  - %s\n", entry.Name())
	}

	// Print initial report
	fmt.Printf("\n%s\n", af.GetReport())
}

// ListRules returns all rule names from testdata directory
func ListRules() []string {
	testdataDir := filepath.Join("rules", "testdata")
	entries, _ := os.ReadDir(testdataDir)

	var rules []string
	for _, entry := range entries {
		if entry.IsDir() {
			rules = append(rules, entry.Name())
		}
	}

	return rules
}
