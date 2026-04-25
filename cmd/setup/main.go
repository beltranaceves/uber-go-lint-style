// Package main implements a small interactive setup helper that writes
// configuration files to help install and run the `uber-go-lint-style`
// golangci-lint plugin.
//
// # Diagnostics and environment flags
//
//   - `SETUP_VERBOSE`: when non-empty the program prints extra diagnostic
//     information (detected release, dry-run previews, and merge diagnostics).
//
//   - `SETUP_DRY_RUN`: when non-empty the program will not write files to
//     disk; instead it prints what it would create/merge. Useful for
//     reproducing the exact YAML that would be written.
//
// # Version resolution
//
// The setup tool attempts to pin the plugin `version:` written into
// generated YAML so that what it writes matches what `go run
// github.com/.../cmd/setup@latest` would actually resolve. To do that it
// first invokes `go list -m -json <module>@latest` (this mirrors the
// Go toolchain / module proxy resolution). If that fails it falls back to
// querying the GitHub Releases API (`/repos/:owner/:repo/releases/latest`).
//
// Notes / debugging tips:
//   - To force the Go toolchain to fetch from the VCS instead of a proxy,
//     run with `GOPROXY=direct`.
//   - To preview what would be written without changing files run:
//     `SETUP_VERBOSE=1 SETUP_DRY_RUN=1 go run ./cmd/setup`
//   - The program currently does not consume `GITHUB_TOKEN`; setting it may
//     increase API limits but is not wired in yet.
//
// These diagnostics were added because `go run <module>@latest` and
// GitHub "releases" can diverge when proxies or tag naming differs; the
// dual-check approach improves the chance the file contains the version
// that `go` will actually use.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const customGclConfig = `version: v2.11.4

plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    import: 'github.com/beltranaceves/uber-go-lint-style'
    version: 'latest'
`

const golangciConfig = `version: "2"

linters:
  default: none
  enable:
    - uber-go-lint-style
  settings:
    custom:
      uber-go-lint-style:
        type: "module"
        description: "Uber Go style guide linter"
        original-url: "github.com/beltranaceves/uber-go-lint-style"
        # Disabled rules provided as YAML text. By default exclude TodoRule.
        settings:
          disabled_rules_yaml: |
            - todo

severity:
  default: info
  rules:
    - linters:
        - uber-go-lint-style
      severity: warning
`

const makefile = `

.PHONY: uber_lint
uber_lint: # Run Uber Go style linter (builds plugin if needed)
	$Q echo "Running Uber Go style linter (with golangci-lint)..."
	$Q if [ ! -f "./custom-gcl" ]; then echo "Building custom golangci-lint with uber-go-lint-style plugin..."; golangci-lint custom || exit 1; fi; echo "Running Uber Go style golangci-lint..." ;./custom-gcl run --config .golangci.uber_style.yml

.PHONY: uber_clean
uber_clean: # Clean Uber Go style linter artifacts
	$Q rm -f custom-gcl*
	$Q echo "Cleaned Uber Go style linter artifacts"
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
	fmt.Println("     (The first time it takes ~1-2 minutes to build the plugin)")
	fmt.Println("")
	fmt.Println("  2. View results:")
	fmt.Println("     Rule violations will be reported in your code")
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
	// Determine latest plugin release (fallback to 'latest' when unavailable)
	release := getLatestReleaseVersion("beltranaceves", "uber-go-lint-style")
	if isVerbose() {
		if release == "" {
			fmt.Printf("  ℹ️  Detected release: <empty> (will use 'latest' literal)\n")
		} else {
			fmt.Printf("  ℹ️  Detected release: %s\n", release)
		}
	}

	// Create YAML config files with interactive prompts
	// If the repo already contains any common golangci config filename
	// prefer prompting on that file so the user gets offered a merge.
	golangciNames := []string{".golangci.uber_style.yml", ".golangci.yml", "golangci.yml", ".golangci.yaml", "golangci.yaml"}
	chosenGolangci := ".golangci.uber_style.yml"
	for _, n := range golangciNames {
		if _, err := os.Stat(n); err == nil {
			chosenGolangci = n
			break
		}
	}

	// Replace plugin version placeholders with the discovered release tag when available
	customGcl := customGclConfig
	golangciCfg := golangciConfig
	if release != "" {
		customGcl = strings.Replace(customGcl, "version: 'latest'", fmt.Sprintf("version: '%s'", release), 1)
		golangciCfg = strings.Replace(golangciCfg, "version: 'latest'", fmt.Sprintf("version: '%s'", release), 1)
	}

	yamlFiles := map[string]string{
		".custom-gcl.yml": customGcl,
		chosenGolangci:    golangciCfg,
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

// getLatestReleaseVersion queries the GitHub Releases API for the latest tag.
// Returns empty string on error (caller should fallback to 'latest').
func getLatestReleaseVersion(owner, repo string) string {
	// Prefer Go's module resolution (what `go run ...@latest` will actually use).
	modulePath := fmt.Sprintf("github.com/%s/%s", owner, repo)
	if v := getModuleVersionViaGoList(modulePath); v != "" {
		return v
	}

	// Fallback to GitHub Releases API when `go list` is unavailable.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ""
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var data struct {
		TagName string `json:"tag_name"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&data); err != nil {
		return ""
	}

	return strings.TrimSpace(data.TagName)
}

// getModuleVersionViaGoList invokes `go list -m -json <module>@latest` and
// returns the resolved Version (or empty string on error). This mirrors the
// version `go run <module>@latest` will pick via GOPROXY/module proxy.
func getModuleVersionViaGoList(module string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-json", module+"@latest")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	var info struct {
		Version string `json:"Version"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return ""
	}
	return strings.TrimSpace(info.Version)
}

// createOrUpdateFile handles creation and updating of files with user interaction.
// isYAML indicates if collision detection should attempt YAML parsing.
func createOrUpdateFile(filename, content string, isYAML bool) error {
	dry := isDryRun()
	verbose := isVerbose()

	existingContent, err := os.ReadFile(filename)
	if err != nil {
		// File doesn't exist, create it
		if dry {
			if verbose {
				fmt.Printf("  DRY-RUN: would create %s with content:\n%s\n", filename, indent(content, "    "))
			} else {
				fmt.Printf("  DRY-RUN: would create %s\n", filename)
			}
			return nil
		}

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

			if dry {
				if verbose {
					fmt.Printf("  DRY-RUN: would merge into %s resulting content:\n%s\n", filename, indent(merged, "    "))
				} else {
					fmt.Printf("  DRY-RUN: would merge into %s\n", filename)
				}
				return nil
			}

			if err := os.WriteFile(filename, []byte(merged), 0644); err != nil {
				return fmt.Errorf("failed to merge %s: %w", filename, err)
			}
			fmt.Printf("  ✓ Merged uber-go-lint-style settings into %s\n", filename)
			return nil

		case "overwrite":
			if dry {
				if verbose {
					fmt.Printf("  DRY-RUN: would overwrite %s with:\n%s\n", filename, indent(content, "    "))
				} else {
					fmt.Printf("  DRY-RUN: would overwrite %s\n", filename)
				}
				return nil
			}
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
			if dry {
				if verbose {
					fmt.Printf("  DRY-RUN: would overwrite %s with:\n%s\n", filename, indent(content, "    "))
				} else {
					fmt.Printf("  DRY-RUN: would overwrite %s\n", filename)
				}
				return nil
			}
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
		if dry {
			if verbose {
				fmt.Printf("  DRY-RUN: would overwrite %s with:\n%s\n", filename, indent(content, "    "))
			} else {
				fmt.Printf("  DRY-RUN: would overwrite %s\n", filename)
			}
			return nil
		}
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
	// Support old top-level `linters-settings.custom` shape
	if lintersSettings, ok := pluginCfg["linters-settings"].(map[string]any); ok {
		if custom, ok := lintersSettings["custom"].(map[string]any); ok {
			if ruleCfg, ok := custom["uber-go-lint-style"].(map[string]any); ok {
				return ruleCfg
			}
		}
	}

	// Support nested `linters.settings.custom` shape (golangci-lint v2 style)
	if linters, ok := pluginCfg["linters"].(map[string]any); ok {
		if settings, ok := linters["settings"].(map[string]any); ok {
			if custom, ok := settings["custom"].(map[string]any); ok {
				if ruleCfg, ok := custom["uber-go-lint-style"].(map[string]any); ok {
					return ruleCfg
				}
			}
		}
	}

	return nil
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

	// getLatestReleaseVersion queries the GitHub Releases API for the latest tag.
	// Returns empty string on error (caller should fallback to 'latest').

	for _, plugin := range cfg.Plugins {
		if strings.Contains(plugin.Module, "uber-go-lint-style") {
			if v := strings.TrimSpace(plugin.Version); v != "" {
				return v
			}
		}
	}

	// Fallback: scan text for a module line and nearby version line (more tolerant)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "uber-go-lint-style") && strings.Contains(line, "module") {
			// look ahead a few lines for a version field
			for j := i; j < i+6 && j < len(lines); j++ {
				if strings.Contains(lines[j], "version:") {
					parts := strings.SplitN(lines[j], ":", 2)
					if len(parts) < 2 {
						continue
					}
					val := strings.TrimSpace(parts[1])
					val = strings.Trim(val, " '\t\"`")
					if val != "" {
						return val
					}
				}
			}
		}
	}

	return ""
}

// isGolangciConfig returns true for common golangci config filenames.
func isGolangciConfig(name string) bool {
	b := filepath.Base(name)
	b = strings.TrimPrefix(b, ".")
	return strings.HasPrefix(b, "golangci")
}

// promptForAction asks the user to choose an action for file handling.
func promptForAction(filename string, options ...string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Build display like: [(M)erge, (s)kip, (o)verwrite]
		disp := make([]string, 0, len(options))
		for i, opt := range options {
			if opt == "" {
				continue
			}
			runes := []rune(opt)
			if len(runes) == 0 {
				continue
			}
			first := string(runes[0])
			rest := ""
			if len(runes) > 1 {
				rest = string(runes[1:])
			}
			if i == 0 {
				// Default option: capitalise the shorthand
				disp = append(disp, fmt.Sprintf("(%s)%s", strings.ToUpper(first), rest))
			} else {
				disp = append(disp, fmt.Sprintf("(%s)%s", strings.ToLower(first), rest))
			}
		}

		fmt.Printf("  Options: %s: [%s]: ", filename, strings.Join(disp, ", "))

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		lower := strings.ToLower(input)

		// Default to first option if empty
		if lower == "" {
			return options[0]
		}

		// Check if input matches any option (by first letter or full name)
		for _, opt := range options {
			if opt == "" {
				continue
			}
			optLower := strings.ToLower(opt)
			firstLower := strings.ToLower(string([]rune(opt)[0]))
			if lower == optLower || lower == firstLower {
				return opt
			}
		}

		fmt.Printf("  Invalid choice. Please enter one of: %s\n", strings.Join(options, ", "))
	}
}

func createOrMergeMakefile() error {
	const makefileName = "Makefile"

	if isVerbose() {
		fmt.Printf("  ℹ️  createOrMergeMakefile verbose enabled\n")
	}

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

		if isDryRun() {
			if isVerbose() {
				fmt.Printf("  DRY-RUN: would merge into %s resulting content:\n%s\n", makefileName, indent(mergedContent, "    "))
			} else {
				fmt.Printf("  DRY-RUN: would merge into %s\n", makefileName)
			}
			return nil
		}

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

func isVerbose() bool {
	return os.Getenv("SETUP_VERBOSE") != ""
}

func isDryRun() bool {
	return os.Getenv("SETUP_DRY_RUN") != ""
}
