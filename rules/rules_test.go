package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mgechev/revive/lint"
)

func isVerboseMode() bool {
	for _, arg := range os.Args {
		if arg == "--verbose" {
			return true
		}
	}
	return false
}

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
	verbose := isVerboseMode()

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

	if verbose {
		fmt.Printf("Found %d rules to test\n", len(rules))
	}

	for _, ruleName := range rules {
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

// testRule tests a single rule against its test cases
func testRule(t *testing.T, ruleName string) {
	testdataDir := filepath.Join(getTestdataDir(), ruleName)
	ruleFile := filepath.Join(filepath.Dir(getTestdataDir()), ruleName+".go")

	if _, err := os.Stat(ruleFile); err != nil {
		t.Fatalf("Rule file not found for %s: %v", ruleName, err)
	}

	newRule, ok := ruleFactory(ruleName)
	if !ok {
		t.Fatalf("Rule %s exists but is not registered in test runner", ruleName)
	}

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

	if len(positiveFiles) == 0 || len(negativeFiles) == 0 {
		t.Fatalf("Rule %s must have both positive and negative fixtures", ruleName)
	}

	if isVerboseMode() {
		fmt.Printf("Rule %s: %d positive, %d negative test cases\n",
			ruleName, len(positiveFiles), len(negativeFiles))
	}

	for _, fixture := range positiveFiles {
		fixture := fixture
		t.Run("positive/"+filepath.Base(fixture), func(t *testing.T) {
			t.Parallel()
			failures := runRuleOnFixture(t, newRule(), fixture)
			if len(failures) == 0 {
				t.Fatalf("Expected at least one lint failure for positive fixture: %s", fixture)
			}
		})
	}

	for _, fixture := range negativeFiles {
		fixture := fixture
		t.Run("negative/"+filepath.Base(fixture), func(t *testing.T) {
			t.Parallel()
			failures := runRuleOnFixture(t, newRule(), fixture)
			if len(failures) != 0 {
				t.Fatalf("Expected no lint failures for negative fixture: %s (got %d)", fixture, len(failures))
			}
		})
	}

	if isVerboseMode() {
		fmt.Printf("Rule %s: lint execution completed\n", ruleName)
	}
}

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

func ruleFactory(name string) (func() lint.Rule, bool) {
	factories := map[string]func() lint.Rule{
		"atomic":               func() lint.Rule { return &AtomicRule{} },
		"builtin-name":         func() lint.Rule { return &BuiltinNameRule{} },
		"channel-size":         func() lint.Rule { return &ChannelSizeRule{} },
		"consistency":          func() lint.Rule { return &ConsistencyRule{} },
		"container-capacity":   func() lint.Rule { return &ContainerCapacityRule{} },
		"container-copy":       func() lint.Rule { return &ContainerCopyRule{} },
		"decl-group":           func() lint.Rule { return &DeclGroupRule{} },
		"defer-clean":          func() lint.Rule { return &DeferCleanRule{} },
		"else-unnecessary":     func() lint.Rule { return &ElseUnnecessaryRule{} },
		"embed-public":         func() lint.Rule { return &EmbedPublicRule{} },
		"enum-start":           func() lint.Rule { return &EnumStartRule{} },
		"error-name":           func() lint.Rule { return &ErrorNameRule{} },
		"error-once":           func() lint.Rule { return &ErrorOnceRule{} },
		"error-type":           func() lint.Rule { return &ErrorTypeRule{} },
		"error-wrap":           func() lint.Rule { return &ErrorWrapRule{} },
		"exit-main":            func() lint.Rule { return &ExitMainRule{} },
		"exit-once":            func() lint.Rule { return &ExitOnceRule{} },
		"function-name":        func() lint.Rule { return &FunctionNameRule{} },
		"function-order":       func() lint.Rule { return &FunctionOrderRule{} },
		"functional-option":    func() lint.Rule { return &FunctionalOptionRule{} },
		"global-decl":          func() lint.Rule { return &GlobalDeclRule{} },
		"global-mut":           func() lint.Rule { return &GlobalMutRule{} },
		"global-name":          func() lint.Rule { return &GlobalNameRule{} },
		"goroutine-exit":       func() lint.Rule { return &GoroutineExitRule{} },
		"goroutine-forget":     func() lint.Rule { return &GoroutineForgetRule{} },
		"goroutine-init":       func() lint.Rule { return &GoroutineInitRule{} },
		"import-alias":         func() lint.Rule { return &ImportAliasRule{} },
		"import-group":         func() lint.Rule { return &ImportGroupRule{} },
		"init":                 func() lint.Rule { return &InitRule{} },
		"interface-compliance": func() lint.Rule { return &InterfaceComplianceRule{} },
		"interface-pointer":    func() lint.Rule { return &InterfacePointerRule{} },
		"interface-receiver":   func() lint.Rule { return &InterfaceReceiverRule{} },
		"line-length":          func() lint.Rule { return &LineLengthRule{} },
		"lint":                 func() lint.Rule { return &LintRule{} },
		"map-init":             func() lint.Rule { return &MapInitRule{} },
		"mutex-zero-value":     func() lint.Rule { return &MutexZeroValueRule{} },
		"nest-less":            func() lint.Rule { return &NestLessRule{} },
		"package-name":         func() lint.Rule { return &PackageNameRule{} },
		"panic":                func() lint.Rule { return &PanicRule{} },
		"param-naked":          func() lint.Rule { return &ParamNakedRule{} },
		"performance":          func() lint.Rule { return &PerformanceRule{} },
		"printf-const":         func() lint.Rule { return &PrintfConstRule{} },
		"printf-name":          func() lint.Rule { return &PrintfNameRule{} },
		"slice-nil":            func() lint.Rule { return &SliceNilRule{} },
		"strconv":              func() lint.Rule { return &StrconvRule{} },
		"string-byte-slice":    func() lint.Rule { return &StringByteSliceRule{} },
		"string-escape":        func() lint.Rule { return &StringEscapeRule{} },
		"struct-embed":         func() lint.Rule { return &StructEmbedRule{} },
		"struct-field-key":     func() lint.Rule { return &StructFieldKeyRule{} },
		"struct-field-zero":    func() lint.Rule { return &StructFieldZeroRule{} },
		"struct-pointer":       func() lint.Rule { return &StructPointerRule{} },
		"struct-tag":           func() lint.Rule { return &StructTagRule{} },
		"struct-zero":          func() lint.Rule { return &StructZeroRule{} },
		"test-table":           func() lint.Rule { return &TestTableRule{} },
		"time":                 func() lint.Rule { return &TimeRule{} },
		"type-assert":          func() lint.Rule { return &TypeAssertRule{} },
		"var-decl":             func() lint.Rule { return &VarDeclRule{} },
		"var-scope":            func() lint.Rule { return &VarScopeRule{} },
	}

	f, ok := factories[name]
	return f, ok
}
