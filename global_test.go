package main_test

import (
	"os"
	"os/exec"
	"testing"
)

// This file makes `go test` run the full test suite including rules

func TestFullSuite(t *testing.T) {
	verbose := false
	for _, arg := range os.Args {
		if arg == "--verbose" {
			verbose = true
			break
		}
	}

	// Run all tests including the rules package
	args := []string{"test", "./rules/...", "-run", "TestAllRules"}
	if verbose {
		args = append(args, "-v", "-args", "--verbose")
	}

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "."
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Rules tests failed: %v", err)
	}
}
