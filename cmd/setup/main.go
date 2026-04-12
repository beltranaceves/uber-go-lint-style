package main

import (
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
	fmt.Println("  1. Run: make")
	fmt.Println("     (First time takes ~1-2 minutes to build plugin)")
	fmt.Println("")
	fmt.Println("  2. View results:")
	fmt.Println("     Violations will be reported in your code")
	fmt.Println("")
	fmt.Println("For more info:")
	fmt.Println("  make help")
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
	// Create YAML config files (skip if they exist)
	yamlFiles := map[string]string{
		".custom-gcl.yml": customGclConfig,
		".golangci.yml":   golangciConfig,
	}

	for filename, content := range yamlFiles {
		if _, err := os.Stat(filename); err == nil {
			fmt.Printf("  ℹ️  %s already exists (skipping)\n", filename)
			continue
		}

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
		fmt.Printf("  ✓ Created %s\n", filename)
	}

	// Handle Makefile specially - merge if it exists
	if err := createOrMergeMakefile(); err != nil {
		return err
	}

	return nil
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

	// File exists - check if our lint target is already there
	existingContent := string(content)
	if strings.Contains(existingContent, "lint:") {
		fmt.Printf("  ℹ️  %s already exists with 'lint' target (skipping)\n", makefileName)
		return nil
	}

	// Makefile exists but doesn't have our lint target - try to merge
	fmt.Printf("  ℹ️  %s already exists, merging targets...\n", makefileName)

	// Append our targets with a separator
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
}
