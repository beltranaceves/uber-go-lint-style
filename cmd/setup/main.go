package main

import (
	"fmt"
	"os"
	"os/exec"
)

const customGclConfig = `version: v1.59.0

plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.0
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
`

const makefile = `.DEFAULT_GOAL := lint

# Run linter (builds plugin if needed)
lint:
	@if [ ! -f "./custom-gcl" ]; then \
		echo "Building custom golangci-lint with uber-go-lint-style plugin..."; \
		golangci-lint custom || exit 1; \
	fi
	@./custom-gcl run

# View help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  make             Build plugin (if needed) and run linter"
	@echo "  make clean       Remove cached plugin binary"
	@echo ""
	@echo "Examples:"
	@echo "  make             # First run builds plugin, subsequent runs are fast"
	@echo "  make clean       # Reset and rebuild plugin next time"

.PHONY: clean
clean:
	@rm -f custom-gcl*
	@echo "Cleaned custom linter artifacts"
`

func main() {
	fmt.Println("Setting up uber-go-lint-style plugin...\n")

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
	fmt.Println("✅ Setup complete!\n")
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
	files := map[string]string{
		".custom-gcl.yml": customGclConfig,
		".golangci.yml":   golangciConfig,
		"Makefile":        makefile,
	}

	for filename, content := range files {
		// Check if file exists
		if _, err := os.Stat(filename); err == nil {
			fmt.Printf("  ℹ️  %s already exists (skipping)\n", filename)
			continue
		}

		// Write file
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
		fmt.Printf("  ✓ Created %s\n", filename)
	}

	return nil
}
