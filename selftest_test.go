package main_test

import (
	"os"
	"os/exec"
	"testing"
)

// This file makes `go test` run the full test suite including rules

func TestFullSuite(t *testing.T) {
	// Run all tests including the rules package
	cmd := exec.Command("go", "test", "./rules/...", "-v", "-run", "TestAllRules")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "."
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Rules tests failed: %v", err)
	}
}
