package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadTestData loads all Go files from a testdata directory
func LoadTestData(ruleName string) (positiveFiles, negativeFiles []string, err error) {
	testdataDir := filepath.Join("rules", "testdata", ruleName)

	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading testdata dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		path := filepath.Join(testdataDir, name)
		if strings.HasPrefix(name, "positive") {
			positiveFiles = append(positiveFiles, path)
		} else if strings.HasPrefix(name, "negative") {
			negativeFiles = append(negativeFiles, path)
		}
	}

	return positiveFiles, negativeFiles, nil
}

// ListAllRules returns all available rule names
func ListAllRules() []string {
	testdataDir := filepath.Join("rules", "testdata")

	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		return nil
	}

	var rules []string
	for _, entry := range entries {
		if entry.IsDir() {
			rules = append(rules, entry.Name())
		}
	}

	return rules
}
