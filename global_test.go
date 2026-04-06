package main_test

import (
	"flag"
	"os"
	"os/exec"
	"testing"
)

var verboseMode = flag.Bool("verbose", false, "enable detailed full-suite output")

// This file makes `go test` run the full test suite including rules

func TestFullSuite(t *testing.T) {
	// Run all tests including the rules package
	args := []string{"test", "./rules/...", "-run", "TestAllRules"}
	if verboseMode != nil && *verboseMode {
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
