package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/beltranaceves/uber-go-lint-style/uber_style_revive/internal/linter"
)

var (
	format = flag.String("format", "friendly", "output format: friendly, simple, json")
	list   = flag.Bool("list", false, "list all available rules")
	help   = flag.Bool("help", false, "show help message")
)

const usage = `uber-go-lint - Uber Go Style Guide Linter

Usage:
  uber-go-lint [options] [./path/.../file.go]

Options:
  -format string      Output format: friendly, simple, json (default: friendly)
  -list               List all available rules
  -help               Show this help message

Examples:
  # Lint current directory
  uber-go-lint ./...

  # Lint specific file
  uber-go-lint main.go

  # JSON output
  uber-go-lint -format json ./...

  # List available rules
  uber-go-lint -list
`

func main() {
	flag.Parse()

	if *help {
		fmt.Println(usage)
		return
	}

	if *list {
		listRules()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Error: no paths provided")
		fmt.Println(usage)
		os.Exit(1)
	}

	// Run linter
	runner := linter.NewRunner()
	failures, err := runner.LintPaths(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Format and print results
	output := linter.FormatFailures(failures, *format)
	if output != "" {
		fmt.Print(output)
	}

	// Exit with error code if failures found
	if len(failures) > 0 {
		fmt.Fprintf(os.Stderr, "Found %d lint issues\n", len(failures))
		os.Exit(1)
	}

	fmt.Println("No issues found")
}

func listRules() {
	fmt.Println("Available Uber Go Style Rules:")
	fmt.Println()

	rules := []struct {
		name        string
		description string
	}{
		{"atomic", "Enforce go.uber.org/atomic over sync/atomic"},
		{"error-wrap", "Enforce error wrapping with context"},
		{"error-name", "Enforce standard error variable naming"},
		{"struct-embed", "Prevent embedding of basic types without named fields"},
		{"global-mut", "Discourage mutable global variables"},
	}

	for _, rule := range rules {
		fmt.Printf("  %-15s %s\n", rule.name, rule.description)
	}
}
