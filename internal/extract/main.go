package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RuleExample represents a code example from the style guide
type RuleExample struct {
	RuleName string
	Code     string
	Type     string // "positive" (Bad) or "negative" (Good)
	Column   int    // 0 = Bad column, 1 = Good column
}

func main() {
	rules := []string{
		"interface-pointer", "interface-compliance", "interface-receiver",
		"mutex-zero-value", "container-copy", "defer-clean", "channel-size",
		"enum-start", "time", "type-assert", "panic", "atomic", "global-mut",
		"embed-public", "builtin-name", "init", "exit-main", "exit-once",
		"struct-tag", "goroutine-forget", "goroutine-exit", "goroutine-init",
		"strconv", "string-byte-slice", "container-capacity",
		"line-length", "consistency", "decl-group", "import-group",
		"package-name", "function-name", "import-alias", "function-order",
		"nest-less", "else-unnecessary", "global-decl", "global-name",
		"struct-embed", "var-decl", "slice-nil", "var-scope", "param-naked",
		"string-escape", "struct-field-key", "struct-field-zero", "struct-zero",
		"struct-pointer", "map-init", "printf-const", "printf-name",
		"test-table", "functional-option", "lint",
		"error-type", "error-wrap", "error-name", "error-once",
		"performance",
	}

	rulesDir := "style_guide/rules"
	testdataDir := "rules/testdata"

	os.MkdirAll(testdataDir, 0755)

	for _, ruleName := range rules {
		mdFile := filepath.Join(rulesDir, ruleName+".md")
		if _, err := os.Stat(mdFile); os.IsNotExist(err) {
			fmt.Printf("Skipping %s (file not found)\n", ruleName)
			continue
		}

		content, err := os.ReadFile(mdFile)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", ruleName, err)
			continue
		}

		positiveExamples, negativeExamples := extractExamplesTable(string(content), ruleName)

		ruleDir := filepath.Join(testdataDir, ruleName)
		os.MkdirAll(ruleDir, 0755)

		if len(positiveExamples) > 0 {
			saveExamples(ruleDir, "positive_test.go", positiveExamples)
		}

		if len(negativeExamples) > 0 {
			saveExamples(ruleDir, "negative_test.go", negativeExamples)
		}

		fmt.Printf("Processed %s: %d positive, %d negative\n",
			ruleName, len(positiveExamples), len(negativeExamples))
	}
}

// extractExamplesTable parses markdown tables with Bad/Good columns
func extractExamplesTable(markdown string, ruleName string) (positive, negative []RuleExample) {
	// Find all tables with Bad/Good headers
	tableRE := regexp.MustCompile(`(?s)<table>.*?</table>`)
	tables := tableRE.FindAllString(markdown, -1)

	for _, table := range tables {
		// Find header row to determine column positions
		badHeader := regexp.MustCompile(`(?i)<th>Bad</th>`)
		goodHeader := regexp.MustCompile(`(?i)<th>Good</th>`)

		hasBad := badHeader.FindString(table) != ""
		hasGood := goodHeader.FindString(table) != ""

		if !hasBad && !hasGood {
			continue
		}

		// Parse body rows (skip header row with <th>)
		rowRE := regexp.MustCompile(`(?s)<tr>(.*?)</tr>`)
		rows := rowRE.FindAllStringSubmatch(table, -1)

		// Skip header row (first row with <th>)
		for i := 1; i < len(rows); i++ {
			row := rows[i][1] // Get the row content

			// Find all code blocks in this row - each in a <td> cell
			cellRE := regexp.MustCompile(`(?s)<td>(.*?)</td>`)
			cells := cellRE.FindAllStringSubmatch(row, -1)

			for colIdx, cell := range cells {
				cellContent := cell[1]
				code := extractCodeFromCell(cellContent)
				if code == "" {
					continue
				}

				example := RuleExample{
					RuleName: ruleName,
					Code:     code,
					Column:   colIdx,
				}

				// First column (colIdx == 0) is Bad -> positive
				// Second column (colIdx == 1) is Good -> negative
				if hasBad && colIdx == 0 {
					example.Type = "positive"
					positive = append(positive, example)
				} else if hasGood && colIdx == 1 {
					example.Type = "negative"
					negative = append(negative, example)
				}
			}
		}
	}

	// If no table examples found, try simple extraction
	if len(positive) == 0 && len(negative) == 0 {
		return extractExamplesSimple(markdown, ruleName)
	}

	return positive, negative
}

func extractCodeFromCell(cell string) string {
	// Extract code from ```go ... ``` block within a td
	codeRE := regexp.MustCompile("```go\\s*([\\s\\S]*?)\\s*```")
	match := codeRE.FindStringSubmatch(cell)
	if match == nil {
		return ""
	}
	return strings.TrimSpace(match[1])
}

// Fallback: simple extraction based on section headers
func extractExamplesSimple(markdown string, ruleName string) (positive, negative []RuleExample) {
	codeBlockRE := regexp.MustCompile("```go\\s*([\\s\\S]*?)\\s*```")
	matches := codeBlockRE.FindAllStringSubmatch(markdown, -1)

	// Find all Bad/Good header positions
	badHeaderRE := regexp.MustCompile(`(?i)<th>Bad</th>`)
	goodHeaderRE := regexp.MustCompile(`(?i)<th>Good</th>`)

	badMatches := badHeaderRE.FindAllStringIndex(markdown, -1)
	goodMatches := goodHeaderRE.FindAllStringIndex(markdown, -1)
	codePositions := codeBlockRE.FindAllStringIndex(markdown, -1)

	for i, match := range matches {
		if i >= len(codePositions) {
			break
		}

		code := strings.TrimSpace(match[1])
		// Skip empty or very short code blocks
		if code == "" || len(code) < 10 {
			continue
		}

		pos := codePositions[i][0]

		// Count Bad/Good headers before this code block
		badCount := 0
		goodCount := 0

		for _, bp := range badMatches {
			if bp[0] < pos {
				badCount++
			}
		}
		for _, gp := range goodMatches {
			if gp[0] < pos {
				goodCount++
			}
		}

		example := RuleExample{
			RuleName: ruleName,
			Code:     code,
		}

		// More Bad headers before this code = Bad column (positive)
		// More Good headers before this code = Good column (negative)
		if badCount > goodCount {
			example.Type = "positive"
			positive = append(positive, example)
		} else if goodCount > badCount {
			example.Type = "negative"
			negative = append(negative, example)
		} else if badCount > 0 {
			// Equal but at least one Bad - default to positive
			example.Type = "positive"
			positive = append(positive, example)
		}
	}

	return positive, negative
}

func saveExamples(dir, filename string, examples []RuleExample) {
	var buf bytes.Buffer

	buf.WriteString("// Auto-generated test cases for rule\n")
	buf.WriteString("// Positive = should FAIL lint (Bad code)\n")
	buf.WriteString("// Negative = should PASS lint (Good code)\n\n")
	buf.WriteString("package testdata\n\n")

	for i, ex := range examples {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(fmt.Sprintf("// Example %d\n", i+1))
		buf.WriteString(ex.Code)
		buf.WriteString("\n")
	}

	filepath := filepath.Join(dir, filename)
	os.WriteFile(filepath, buf.Bytes(), 0644)
}
