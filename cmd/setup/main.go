package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const customGclConfig = `version: v1.59.0

plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.1
`

const golangciConfig = `version: "1"

linters:
  disable-all: true
  enable:
    - uber-go-lint-style

linters-settings:
  custom:
    uber-go-lint-style:
      type: "module"
      description: "Uber Go style guide linter"
      original-url: "github.com/beltranaceves/uber-go-lint-style"

severity:
  default-severity: error
  rules:
    - linters:
        - uber-go-lint-style
      severity: warning
`

const makefile = `.DEFAULT_GOAL := uber_lint

# Run linter (builds plugin if needed)
uber_lint:
	@if [ ! -f "./custom-gcl" ]; then \
		echo "Building custom golangci-lint with uber-go-lint-style plugin..."; \
		golangci-lint custom || exit 1; \
	fi
	@./custom-gcl run

# View help
uber_help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  make uber_lint       Build plugin (if needed) and run linter"
	@echo "  make uber_clean      Remove cached plugin binary"
	@echo ""
	@echo "Examples:"
	@echo "  make uber_lint       # First run builds plugin, subsequent runs are fast"
	@echo "  make uber_clean      # Reset and rebuild plugin next time"

.PHONY: uber_lint uber_help uber_clean
uber_clean:
	@rm -f custom-gcl*
	@echo "Cleaned custom linter artifacts"
`

func main() {
	fmt.Println("Setting up uber-go-lint-style plugin...")

	// Check if golangci-lint is installed
	if err := checkGolangciLint(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Create config files
	if err := createConfigFiles(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error creating config files: %v\n", err)
		os.Exit(1)
	}

	// Print success
	fmt.Println("✅ Setup complete!")
	fmt.Println("Next steps:")
	fmt.Println("  1. Run: make uber_lint")
	fmt.Println("     (First time takes ~1-2 minutes to build plugin)")
	fmt.Println("")
	fmt.Println("  2. View results:")
	fmt.Println("     Violations will be reported in your code")
	fmt.Println("")
	fmt.Println("For more info:")
	fmt.Println("  make uber_help")
}

func checkGolangciLint() error {
	cmd := exec.Command("golangci-lint", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"golangci-lint not found. Install with:\n" +
				"  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
		)
	}
	return nil
}

func createConfigFiles() error {
	// Create YAML config files with interactive prompts
	yamlFiles := map[string]string{
		".custom-gcl.yml": customGclConfig,
		".golangci.yml":   golangciConfig,
	}

	for filename, content := range yamlFiles {
		if err := createOrUpdateFile(filename, content, true); err != nil {
			return err
		}
	}

	// Handle Makefile specially - merge if it exists
	if err := createOrMergeMakefile(); err != nil {
		return err
	}

	return nil
}

// createOrUpdateFile handles creation and updating of files with user interaction.
// isYAML indicates if collision detection should attempt YAML parsing.
func createOrUpdateFile(filename, content string, isYAML bool) error {
	existingContent, err := os.ReadFile(filename)
	if err != nil {
		// File doesn't exist, create it
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
		fmt.Printf("  ✓ Created %s\n", filename)
		return nil
	}

	// File exists - check for conflicts and prompt user
	existingStr := string(existingContent)

	// For YAML files, check for collisions
	if isYAML && hasYAMLCollision(existingStr, content) {
		fmt.Printf("\n⚠️  %s exists with conflicting settings (plugin version mismatch)\n", filename)
		action := promptForAction(filename, "overwrite", "skip", "view")
		switch action {
		case "overwrite":
			return os.WriteFile(filename, []byte(content), 0644)
		case "view":
			fmt.Printf("  Existing content:\n%s\n", indent(existingStr, "    "))
			fmt.Printf("  New content:\n%s\n", indent(content, "    "))
			// Ask again after showing
			action = promptForAction(filename, "overwrite", "skip")
			if action == "overwrite" {
				return os.WriteFile(filename, []byte(content), 0644)
			}
		}
		fmt.Printf("  ℹ️  Skipped %s\n", filename)
		return nil
	}

	// No collision - but file exists, prompt for safety
	fmt.Printf("  ℹ️  %s already exists\n", filename)
	action := promptForAction(filename, "skip", "overwrite")
	if action == "overwrite" {
		return os.WriteFile(filename, []byte(content), 0644)
	}
	fmt.Printf("  ℹ️  Skipped %s\n", filename)
	return nil
}

// hasYAMLCollision detects if the plugin version differs between existing and new YAML.
func hasYAMLCollision(existing, new string) bool {
	// Simple version detection: check if plugin version differs
	existingVersion := extractVersionFromYAML(existing)
	newVersion := extractVersionFromYAML(new)
	return existingVersion != "" && newVersion != "" && existingVersion != newVersion
}

// extractVersionFromYAML extracts the plugin version from YAML content.
func extractVersionFromYAML(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "version:") && strings.Contains(line, "v0.") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// promptForAction asks the user to choose an action for file handling.
func promptForAction(filename string, options ...string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("  Options: %s: ", filename)
		for i, opt := range options {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(opt)
		}
		fmt.Print(" [")
		fmt.Print(strings.ToLower(options[0][:1]))
		fmt.Print("]: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		// Default to first option if empty
		if input == "" {
			return options[0]
		}

		// Check if input matches any option (by first letter or full name)
		for _, opt := range options {
			if input == strings.ToLower(opt) || input == strings.ToLower(opt[:1]) {
				return opt
			}
		}

		fmt.Printf("  Invalid choice. Please enter one of: %s\n", strings.Join(options, ", "))
	}
}

func createOrMergeMakefile() error {
	const makefileName = "Makefile"

	// Check if Makefile exists
	content, err := os.ReadFile(makefileName)
	if err != nil {
		// File doesn't exist, create it
		if err := os.WriteFile(makefileName, []byte(makefile), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", makefileName, err)
		}
		fmt.Printf("  ✓ Created %s\n", makefileName)
		return nil
	}

	existingContent := string(content)

	// Check if our uber_lint target already exists
	if strings.Contains(existingContent, "uber_lint:") {
		fmt.Printf("  ℹ️  %s already contains uber_lint target\n", makefileName)
		action := promptForAction(makefileName, "skip", "overwrite", "view")
		switch action {
		case "view":
			fmt.Printf("  Existing content:\n%s\n", indent(existingContent, "    "))
			fmt.Printf("  New content would add:\n%s\n", indent(makefile, "    "))
			// Ask again after showing
			action = promptForAction(makefileName, "skip", "overwrite")
			if action == "overwrite" {
				return os.WriteFile(makefileName, []byte(makefile), 0644)
			}
		case "overwrite":
			return os.WriteFile(makefileName, []byte(makefile), 0644)
		}
		return nil
	}

	// Makefile exists but doesn't have our uber_lint target - offer merge
	fmt.Printf("  ℹ️  %s exists but missing uber_lint targets\n", makefileName)
	action := promptForAction(makefileName, "merge", "skip", "overwrite")

	switch action {
	case "merge":
		fmt.Printf("  Merging uber-go-lint-style targets into %s...\n", makefileName)
		separator := "\n# uber-go-lint-style plugin targets\n"
		mergedContent := existingContent
		if !strings.HasSuffix(mergedContent, "\n") {
			mergedContent += "\n"
		}
		mergedContent += separator + makefile

		if err := os.WriteFile(makefileName, []byte(mergedContent), 0644); err != nil {
			return fmt.Errorf("failed to merge %s: %w", makefileName, err)
		}
		fmt.Printf("  ✓ Merged lint targets into %s\n", makefileName)
		return nil

	case "overwrite":
		return os.WriteFile(makefileName, []byte(makefile), 0644)

	default:
		fmt.Printf("  ℹ️  Skipped %s\n", makefileName)
		return nil
	}
}

// indent adds leading whitespace to each line of text
func indent(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
