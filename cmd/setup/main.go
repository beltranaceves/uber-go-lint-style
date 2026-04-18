package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const customGclConfig = `version: v1.59.0

plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: 'latest'
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
      # Disabled rules provided as YAML text. By default exclude TodoRule.
      disabled_rules_yaml: |
        - TodoRule

severity:
  default-severity: error
  rules:
    - linters:
        - uber-go-lint-style
      severity: warning
`

const makefile = `

.PHONY: uber_lint
uber_lint: # Run Uber Go style linter (builds plugin if needed)
	$Q if [ ! -f "./custom-gcl" ]; then \
	$Q	echo "Building custom golangci-lint with uber-go-lint-style plugin..."; \
	$Q	golangci-lint custom || exit 1; \
	$Q fi
	$Q ./custom-gcl run

.PHONY: uber_clean
uber_clean:
	$Q rm -f custom-gcl*
	$Q echo "Cleaned custom linter artifacts"
`

func main() {
	fmt.Println("Setting up uber-go-lint-style plugin...")

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
	fmt.Println("✅ Setup complete!")
	fmt.Println("Next steps:")
	fmt.Println("  1. Run: make uber_lint")
	fmt.Println("     (First time takes ~1-2 minutes to build plugin)")
	fmt.Println("")
	fmt.Println("  2. View results:")
	fmt.Println("     Violations will be reported in your code")
	fmt.Println("")
	fmt.Println("For more info:")
	fmt.Println("  make uber_help")
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
	// Create YAML config files with interactive prompts
	// If the repo already contains any common golangci config filename
	// prefer prompting on that file so the user gets offered a merge.
	golangciNames := []string{".golangci.yml", "golangci.yml", ".golangci.yaml", "golangci.yaml"}
	chosenGolangci := ".golangci.yml"
	for _, n := range golangciNames {
		if _, err := os.Stat(n); err == nil {
			chosenGolangci = n
			break
		}
	}

	yamlFiles := map[string]string{
		".custom-gcl.yml": customGclConfig,
		chosenGolangci:    golangciConfig,
	}

	for filename, content := range yamlFiles {
		if err := createOrUpdateFile(filename, content, true); err != nil {
			return err
		}
	}

	// Handle Makefile specially - merge if it exists
	if err := createOrMergeMakefile(); err != nil {
		return err
	}

	return nil
}

// createOrUpdateFile handles creation and updating of files with user interaction.
// isYAML indicates if collision detection should attempt YAML parsing.
func createOrUpdateFile(filename, content string, isYAML bool) error {
	existingContent, err := os.ReadFile(filename)
	if err != nil {
		// File doesn't exist, create it
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
		fmt.Printf("  ✓ Created %s\n", filename)
		return nil
	}

	// File exists - check for conflicts and prompt user
	existingStr := string(existingContent)

	if isGolangciConfig(filename) {
		fmt.Printf("  ℹ️  %s already exists\n", filename)
		action := promptForAction(filename, "merge", "skip", "overwrite", "view")
		switch action {
		case "merge":
			merged, changed, mergeErr := mergeGolangCIConfig(existingStr, content)
			if mergeErr != nil {
				fmt.Printf("  ⚠️  Merge failed: %v\n", mergeErr)
				action = promptForAction(filename, "skip", "overwrite", "view")
				switch action {
				case "overwrite":
					return os.WriteFile(filename, []byte(content), 0644)
				case "view":
					fmt.Printf("  Existing content:\n%s\n", indent(existingStr, "    "))
					fmt.Printf("  New content:\n%s\n", indent(content, "    "))
					action = promptForAction(filename, "skip", "overwrite")
					if action == "overwrite" {
						return os.WriteFile(filename, []byte(content), 0644)
					}
				}
				fmt.Printf("  ℹ️  Skipped %s\n", filename)
				return nil
			}

			if !changed {
				fmt.Printf("  ℹ️  %s already contains required uber-go-lint-style settings\n", filename)
				return nil
			}

			if err := os.WriteFile(filename, []byte(merged), 0644); err != nil {
				return fmt.Errorf("failed to merge %s: %w", filename, err)
			}
			fmt.Printf("  ✓ Merged uber-go-lint-style settings into %s\n", filename)
			return nil

		case "overwrite":
			return os.WriteFile(filename, []byte(content), 0644)

		case "view":
			fmt.Printf("  Existing content:\n%s\n", indent(existingStr, "    "))
			fmt.Printf("  New content:\n%s\n", indent(content, "    "))
			action = promptForAction(filename, "merge", "skip", "overwrite")
			switch action {
			case "merge":
				merged, changed, mergeErr := mergeGolangCIConfig(existingStr, content)
				if mergeErr != nil {
					return fmt.Errorf("failed to merge %s: %w", filename, mergeErr)
				}
				if !changed {
					fmt.Printf("  ℹ️  %s already contains required uber-go-lint-style settings\n", filename)
					return nil
				}
				return os.WriteFile(filename, []byte(merged), 0644)
			case "overwrite":
				return os.WriteFile(filename, []byte(content), 0644)
			}
		}

		fmt.Printf("  ℹ️  Skipped %s\n", filename)
		return nil
	}

	// For YAML files, check for collisions
	if isYAML && hasYAMLCollision(existingStr, content) {
		fmt.Printf("\n⚠️  %s exists with conflicting settings (plugin version mismatch)\n", filename)
		action := promptForAction(filename, "overwrite", "skip", "view")
		switch action {
		case "overwrite":
			return os.WriteFile(filename, []byte(content), 0644)
		case "view":
			fmt.Printf("  Existing content:\n%s\n", indent(existingStr, "    "))
			fmt.Printf("  New content:\n%s\n", indent(content, "    "))
			// Ask again after showing
			action = promptForAction(filename, "overwrite", "skip")
			if action == "overwrite" {
				return os.WriteFile(filename, []byte(content), 0644)
			}
		}
		fmt.Printf("  ℹ️  Skipped %s\n", filename)
		return nil
	}

	// No collision - but file exists, prompt for safety
	fmt.Printf("  ℹ️  %s already exists\n", filename)
	action := promptForAction(filename, "skip", "overwrite")
	if action == "overwrite" {
		return os.WriteFile(filename, []byte(content), 0644)
	}
	fmt.Printf("  ℹ️  Skipped %s\n", filename)
	return nil
}

func mergeGolangCIConfig(existingContent, pluginContent string) (string, bool, error) {
	existingCfg := map[string]any{}
	pluginCfg := map[string]any{}

	if err := yaml.Unmarshal([]byte(existingContent), &existingCfg); err != nil {
		return "", false, fmt.Errorf("parse existing YAML: %w", err)
	}
	if err := yaml.Unmarshal([]byte(pluginContent), &pluginCfg); err != nil {
		return "", false, fmt.Errorf("parse plugin YAML: %w", err)
	}

	changed := false

	if mergeLinters(existingCfg) {
		changed = true
	}
	if mergeLinterSettings(existingCfg, pluginCfg) {
		changed = true
	}
	if mergeSeverityRules(existingCfg) {
		changed = true
	}

	if !changed {
		return existingContent, false, nil
	}

	mergedBytes, err := yaml.Marshal(existingCfg)
	if err != nil {
		return "", false, fmt.Errorf("marshal merged YAML: %w", err)
	}

	return string(mergedBytes), true, nil
}

func mergeLinters(cfg map[string]any) bool {
	linters := ensureMap(cfg, "linters")

	enable, ok := linters["enable"].([]any)
	if !ok {
		enable = []any{}
	}

	if stringSliceContains(enable, "uber-go-lint-style") {
		if _, exists := linters["enable"]; !exists {
			linters["enable"] = enable
		}
		return false
	}

	enable = append(enable, "uber-go-lint-style")
	linters["enable"] = enable
	return true
}

func mergeLinterSettings(existingCfg, pluginCfg map[string]any) bool {
	pluginSettings := getPluginRuleSettings(pluginCfg)
	if pluginSettings == nil {
		return false
	}

	lintersSettings := ensureMap(existingCfg, "linters-settings")
	custom := ensureNestedMap(lintersSettings, "custom")
	ruleCfg, exists := custom["uber-go-lint-style"].(map[string]any)
	if !exists {
		ruleCfg = map[string]any{}
		custom["uber-go-lint-style"] = ruleCfg
	}

	changed := false
	for key, value := range pluginSettings {
		if _, exists := ruleCfg[key]; !exists {
			ruleCfg[key] = value
			changed = true
		}
	}

	return changed
}

func mergeSeverityRules(cfg map[string]any) bool {
	severity := ensureMap(cfg, "severity")
	rules, ok := severity["rules"].([]any)
	if !ok {
		rules = []any{}
	}

	for _, rule := range rules {
		ruleMap, ok := rule.(map[string]any)
		if !ok {
			continue
		}
		linters, ok := ruleMap["linters"].([]any)
		if !ok {
			continue
		}
		if stringSliceContains(linters, "uber-go-lint-style") {
			return false
		}
	}

	severityRule := map[string]any{
		"linters":  []any{"uber-go-lint-style"},
		"severity": "warning",
	}
	severity["rules"] = append(rules, severityRule)
	return true
}

func getPluginRuleSettings(pluginCfg map[string]any) map[string]any {
	lintersSettings, ok := pluginCfg["linters-settings"].(map[string]any)
	if !ok {
		return nil
	}
	custom, ok := lintersSettings["custom"].(map[string]any)
	if !ok {
		return nil
	}
	ruleCfg, ok := custom["uber-go-lint-style"].(map[string]any)
	if !ok {
		return nil
	}
	return ruleCfg
}

func ensureMap(root map[string]any, key string) map[string]any {
	child, ok := root[key].(map[string]any)
	if ok {
		return child
	}
	child = map[string]any{}
	root[key] = child
	return child
}

func ensureNestedMap(root map[string]any, key string) map[string]any {
	child, ok := root[key].(map[string]any)
	if ok {
		return child
	}
	child = map[string]any{}
	root[key] = child
	return child
}

func stringSliceContains(values []any, target string) bool {
	for _, value := range values {
		str, ok := value.(string)
		if ok && str == target {
			return true
		}
	}
	return false
}

// hasYAMLCollision detects if the plugin version differs between existing and new YAML.
func hasYAMLCollision(existing, new string) bool {
	// Simple version detection: check if plugin version differs
	existingVersion := extractVersionFromYAML(existing)
	newVersion := extractVersionFromYAML(new)
	return existingVersion != "" && newVersion != "" && existingVersion != newVersion
}

// extractVersionFromYAML extracts the plugin version from YAML content.
func extractVersionFromYAML(content string) string {
	type pluginEntry struct {
		Module  string `yaml:"module"`
		Version string `yaml:"version"`
	}
	type customGCLConfig struct {
		Plugins []pluginEntry `yaml:"plugins"`
	}

	var cfg customGCLConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return ""
	}

	for _, plugin := range cfg.Plugins {
		if strings.Contains(plugin.Module, "uber-go-lint-style") {
			return strings.TrimSpace(plugin.Version)
		}
	}

	return ""
}

// isGolangciConfig returns true for common golangci config filenames.
func isGolangciConfig(name string) bool {
	b := filepath.Base(name)
	b = strings.TrimPrefix(b, ".")
	return b == "golangci.yml" || b == "golangci.yaml"
}

// promptForAction asks the user to choose an action for file handling.
func promptForAction(filename string, options ...string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("  Options: %s: ", filename)
		for i, opt := range options {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(opt)
		}
		fmt.Print(" [")
		fmt.Print(strings.ToLower(options[0][:1]))
		fmt.Print("]: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		// Default to first option if empty
		if input == "" {
			return options[0]
		}

		// Check if input matches any option (by first letter or full name)
		for _, opt := range options {
			if input == strings.ToLower(opt) || input == strings.ToLower(opt[:1]) {
				return opt
			}
		}

		fmt.Printf("  Invalid choice. Please enter one of: %s\n", strings.Join(options, ", "))
	}
}

func createOrMergeMakefile() error {
	const makefileName = "Makefile"

	// Check if Makefile exists
	content, err := os.ReadFile(makefileName)
	if err != nil {
		// File doesn't exist, create it
		if err := os.WriteFile(makefileName, []byte(makefile), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", makefileName, err)
		}
		fmt.Printf("  ✓ Created %s\n", makefileName)
		return nil
	}

	existingContent := string(content)

	// Check if our uber_lint target already exists
	if strings.Contains(existingContent, "uber_lint:") {
		fmt.Printf("  ℹ️  %s already contains uber_lint target\n", makefileName)
		action := promptForAction(makefileName, "skip", "overwrite", "view")
		switch action {
		case "view":
			fmt.Printf("  Existing content:\n%s\n", indent(existingContent, "    "))
			fmt.Printf("  New content would add:\n%s\n", indent(makefile, "    "))
			// Ask again after showing
			action = promptForAction(makefileName, "skip", "overwrite")
			if action == "overwrite" {
				return os.WriteFile(makefileName, []byte(makefile), 0644)
			}
		case "overwrite":
			return os.WriteFile(makefileName, []byte(makefile), 0644)
		}
		return nil
	}

	// Makefile exists but doesn't have our uber_lint target - offer merge
	fmt.Printf("  ℹ️  %s exists but missing uber_lint targets\n", makefileName)
	action := promptForAction(makefileName, "merge", "skip", "overwrite")

	switch action {
	case "merge":
		fmt.Printf("  Merging uber-go-lint-style targets into %s...\n", makefileName)
		separator := "\n# uber-go-lint-style plugin targets\n"
		mergedContent := existingContent
		if !strings.HasSuffix(mergedContent, "\n") {
			mergedContent += "\n"
		}
		mergedContent += separator + makefile

		if err := os.WriteFile(makefileName, []byte(mergedContent), 0644); err != nil {
			return fmt.Errorf("failed to merge %s: %w", makefileName, err)
		}
		fmt.Printf("  ✓ Merged lint targets into %s\n", makefileName)
		return nil

	case "overwrite":
		return os.WriteFile(makefileName, []byte(makefile), 0644)

	default:
		fmt.Printf("  ℹ️  Skipped %s\n", makefileName)
		return nil
	}
}

// indent adds leading whitespace to each line of text
func indent(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
