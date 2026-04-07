package rules

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mgechev/revive/lint"
)

var verboseMode = flag.Bool("verbose", false, "enable detailed rule test output")

func getTestdataDir() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "..", "testdata")
}

// TestAllRules runs tests for all Uber Go Style rules.
func TestAllRules(t *testing.T) {
	testdataDir := getTestdataDir()
	verbose := *verboseMode

	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	var ruleNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			ruleNames = append(ruleNames, entry.Name())
		}
	}

	if len(ruleNames) == 0 {
		t.Skip("No rules found in testdata directory")
	}

	if verbose {
		fmt.Printf("Found %d rules to test\n", len(ruleNames))
	}

	for _, ruleName := range ruleNames {
		ruleName := ruleName
		passed := t.Run(ruleName, func(t *testing.T) {
			testRule(t, ruleName)
		})

		status := "PASS"
		if !passed {
			status = "FAIL"
		}
		fmt.Printf("rule %-20s %s\n", ruleName, status)
	}
}

// testRule tests a single rule
func testRule(t *testing.T, ruleName string) {
	testdataDir := filepath.Join(getTestdataDir(), ruleName)

	// Verify rule exists
	rule, ok := NewRule(ruleName)
	if !ok {
		t.Fatalf("Rule %s not registered", ruleName)
	}

	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("Failed to read testdata directory for %s: %v", ruleName, err)
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

	if len(positiveFiles) == 0 {
		t.Fatalf("Rule %s: no positive test fixtures found", ruleName)
	}
	if len(negativeFiles) == 0 {
		t.Fatalf("Rule %s: no negative test fixtures found", ruleName)
	}

	// Test positive cases (should have failures)
	for _, fixture := range positiveFiles {
		fixture := fixture
		t.Run("positive/"+filepath.Base(fixture), func(t *testing.T) {
			t.Parallel()
			failures := runRuleOnFixture(t, NewRuleInstance(ruleName), fixture)
			if len(failures) == 0 {
				t.Fatalf("Expected lint failures for positive fixture: %s", filepath.Base(fixture))
			}
		})
	}

	// Test negative cases (should have no failures)
	for _, fixture := range negativeFiles {
		fixture := fixture
		t.Run("negative/"+filepath.Base(fixture), func(t *testing.T) {
			t.Parallel()
			failures := runRuleOnFixture(t, NewRuleInstance(ruleName), fixture)
			if len(failures) != 0 {
				t.Fatalf("Expected no lint failures for negative fixture: %s (got %d)",
					filepath.Base(fixture), len(failures))
			}
		})
	}
}

// NewRuleInstance creates a fresh rule instance (for use in tests)
func NewRuleInstance(name string) lint.Rule {
	rule, _ := NewRule(name)
	return rule
}

// runRuleOnFixture runs a rule on a single test fixture
func runRuleOnFixture(t *testing.T, rule lint.Rule, fixturePath string) []lint.Failure {
	t.Helper()

	linter := lint.New(os.ReadFile, 0)
	config := lint.Config{
		Confidence: 0,
		Rules: lint.RulesConfig{
			rule.Name(): {
				Arguments: nil,
			},
		},
		Directives: lint.DirectivesConfig{},
	}

	failuresCh, err := linter.Lint([][]string{{fixturePath}}, []lint.Rule{rule}, config)
	if err != nil {
		t.Fatalf("Failed to lint fixture %s: %v", fixturePath, err)
	}

	var failures []lint.Failure
	for failure := range failuresCh {
		if failure.IsInternal() {
			t.Fatalf("Internal lint failure for fixture %s: %s", fixturePath, failure.Failure)
		}
		failures = append(failures, failure)
	}

	return failures
}
