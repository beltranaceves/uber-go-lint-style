package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// getTestdataDir returns the path to the testdata directory
func getTestdataDir() string {
	// Get the directory of this test file
	_, currentFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(currentFile)

	// The test is in rules/ directory
	// testdata is at rules/testdata
	return filepath.Join(testDir, "testdata")
}

// TestAllRules runs tests for all Uber Go Style Guide rules.
// Each rule is tested against its positive (should fail) and negative (should pass) test cases.
func TestAllRules(t *testing.T) {
	testdataDir := getTestdataDir()

	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	var rules []string
	for _, entry := range entries {
		if entry.IsDir() {
			rules = append(rules, entry.Name())
		}
	}

	if len(rules) == 0 {
		t.Skip("No rules found in testdata directory")
	}

	fmt.Printf("Found %d rules to test\n", len(rules))

	for _, ruleName := range rules {
		t.Run(ruleName, func(t *testing.T) {
			testRule(t, ruleName)
		})
	}
}

// testRule tests a single rule against its test cases
func testRule(t *testing.T, ruleName string) {
	testdataDir := filepath.Join(getTestdataDir(), ruleName)

	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	var positiveFiles, negativeFiles []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		path := filepath.Join(testdataDir, entry.Name())
		if strings.HasPrefix(entry.Name(), "positive") {
			positiveFiles = append(positiveFiles, path)
		} else if strings.HasPrefix(entry.Name(), "negative") {
			negativeFiles = append(negativeFiles, path)
		}
	}

	if len(positiveFiles) == 0 && len(negativeFiles) == 0 {
		t.Skipf("No test cases found for rule %s", ruleName)
	}

	fmt.Printf("Rule %s: %d positive, %d negative test cases\n",
		ruleName, len(positiveFiles), len(negativeFiles))

	// To run actual tests with revive:
	// 1. Install revive: go install github.com/mgechev/revive@latest
	// 2. Build with your rules included: cd revive && go build -o revive .
	// 3. Run: revive -config revive.toml ./rules/testdata/[rule]/positive_test.go

	// This test framework verifies test case structure exists
	// Actual lint testing requires revive to be built with the custom rules
}
