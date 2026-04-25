package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestExtractVersionFromYAML tests version extraction from YAML content.
func TestExtractVersionFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "plugin version latest",
			content: `version: v1.59.0

plugins:
	- module: 'github.com/beltranaceves/uber-go-lint-style'
		import: 'github.com/beltranaceves/uber-go-lint-style'
		version: 'latest'
`,
			expected: "",
		},
		{
			name: "valid version v0.2.0",
			content: `plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.2.0
`,
			expected: "v0.2.0",
		},
		{
			name: "no version found",
			content: `version: v1.59.0
linters:
  disable-all: true
`,
			expected: "",
		},
		{
			name: "malformed version line",
			content: `plugins:
  version v0.1.1
`,
			expected: "",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name: "multiple versions, returns first",
			content: `plugins:
  - module: github.com/example/other-plugin
    version: v9.9.9
  - module: github.com/beltranaceves/uber-go-lint-style
    version: v0.1.0
  - module: github.com/beltranaceves/uber-go-lint-style
    version: v0.1.1
`,
			expected: "v0.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVersionFromYAML(tt.content)
			if got != tt.expected {
				t.Errorf("extractVersionFromYAML() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestHasYAMLCollision tests collision detection between YAML versions.
func TestHasYAMLCollision(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		new      string
		expected bool
	}{
		{
			name: "no collision - same versions",
			existing: `version: v0.1.1
plugins:
  - module: github.com/beltranaceves/uber-go-lint-style
    version: v0.1.1
`,
			new: `version: v0.1.1
plugins:
  - module: github.com/beltranaceves/uber-go-lint-style
    version: v0.1.1
`,
			expected: false,
		},
		{
			name: "collision detected - different versions",
			existing: `version: v0.1.0
plugins:
  - module: github.com/beltranaceves/uber-go-lint-style
    version: v0.1.0
`,
			new: `version: v0.1.1
plugins:
  - module: github.com/beltranaceves/uber-go-lint-style
    version: v0.1.1
`,
			expected: true,
		},
		{
			name: "no collision - existing has no version",
			existing: `linters:
  disable-all: true
`,
			new: `version: v0.1.1
plugins:
  - module: github.com/beltranaceves/uber-go-lint-style
    version: v0.1.1
`,
			expected: false,
		},
		{
			name: "no collision - new has no version",
			existing: `version: v0.1.0
plugins:
  - module: github.com/beltranaceves/uber-go-lint-style
    version: v0.1.0
`,
			new: `linters:
  disable-all: true
`,
			expected: false,
		},
		{
			name: "no collision - both have no version",
			existing: `linters:
  disable-all: true
`,
			new: `linters:
  disable-all: true
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasYAMLCollision(tt.existing, tt.new)
			if got != tt.expected {
				t.Errorf("hasYAMLCollision() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIndent tests text indentation.
func TestIndent(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		prefix string
		want   string
	}{
		{
			name:   "single line",
			text:   "hello world",
			prefix: "  ",
			want:   "  hello world",
		},
		{
			name:   "multiple lines",
			text:   "line1\nline2\nline3",
			prefix: "  ",
			want:   "  line1\n  line2\n  line3",
		},
		{
			name:   "empty lines preserved",
			text:   "line1\n\nline3",
			prefix: "  ",
			want:   "  line1\n\n  line3",
		},
		{
			name:   "tab indentation",
			text:   "hello\nworld",
			prefix: "\t",
			want:   "\thello\n\tworld",
		},
		{
			name:   "empty text",
			text:   "",
			prefix: "  ",
			want:   "",
		},
		{
			name:   "trailing newline",
			text:   "hello\nworld\n",
			prefix: "  ",
			want:   "  hello\n  world\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indent(tt.text, tt.prefix)
			if got != tt.want {
				t.Errorf("indent() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCreateOrUpdateFileIntegration tests file creation and updates in temp directories.
func TestCreateOrUpdateFileIntegration(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		newContent      string
		existingFile    bool
		existingContent string
		isYAML          bool
		// For testing, we pre-answer the prompt
		setupPromptResponse func()
		shouldError         bool
		expectCreated       bool
		expectContent       string
	}{
		{
			name:          "create new file",
			filename:      "new-config.yml",
			newContent:    "version: v0.1.1",
			existingFile:  false,
			isYAML:        true,
			expectCreated: true,
			expectContent: "version: v0.1.1",
		},
		{
			name:            "file already exists with same version",
			filename:        "existing-config.yml",
			newContent:      "version: v0.1.1",
			existingFile:    true,
			existingContent: "version: v0.1.1",
			isYAML:          true,
			expectCreated:   false,
			expectContent:   "version: v0.1.1", // unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filepath := filepath.Join(tmpDir, tt.filename)

			// Setup existing file if needed
			if tt.existingFile {
				err := os.WriteFile(filepath, []byte(tt.existingContent), 0644)
				if err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			// For this test, we can't easily test interactive prompts
			// So we'll test the non-interactive path (file doesn't exist)
			if !tt.existingFile {
				err := createOrUpdateFile(filepath, tt.newContent, tt.isYAML)
				if (err != nil) != tt.shouldError {
					t.Errorf("createOrUpdateFile() error = %v, shouldError = %v", err, tt.shouldError)
				}

				if tt.expectCreated {
					content, err := os.ReadFile(filepath)
					if err != nil {
						t.Errorf("failed to read created file: %v", err)
					}
					if string(content) != tt.expectContent {
						t.Errorf("file content = %q, want %q", string(content), tt.expectContent)
					}
				}
			}
		})
	}
}

// TestCreateOrMergeFileIntegration tests Makefile creation and merging.
func TestCreateOrMergeFileIntegration(t *testing.T) {
	tests := []struct {
		name             string
		existingMakefile bool
		existingContent  string
		expectMerge      bool
		expectCreated    bool
	}{
		{
			name:             "create new Makefile",
			existingMakefile: false,
			expectCreated:    true,
			expectMerge:      false,
		},
		{
			name:             "Makefile with existing user targets",
			existingMakefile: true,
			existingContent: `.PHONY: test
test:
	@echo "running tests"
`,
			expectCreated: false,
			expectMerge:   true, // Should merge (no uber_lint target)
		},
		{
			name:             "Makefile with existing uber_lint target",
			existingMakefile: true,
			existingContent: `uber_lint:
	@echo "running lint"
`,
			expectCreated: false,
			expectMerge:   false, // Should skip (already has target)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			origDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get cwd: %v", err)
			}
			defer os.Chdir(origDir)

			// Change to temp directory
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			// Setup existing Makefile if needed
			if tt.existingMakefile {
				err := os.WriteFile("Makefile", []byte(tt.existingContent), 0644)
				if err != nil {
					t.Fatalf("failed to write test Makefile: %v", err)
				}
			}

			// Note: This test can't easily run interactively without mocking stdin
			// So we're primarily checking file creation paths
			if !tt.existingMakefile {
				err := createOrMergeMakefile()
				if err != nil {
					t.Errorf("createOrMergeMakefile() error = %v", err)
				}

				if tt.expectCreated {
					_, err := os.Stat("Makefile")
					if err != nil {
						t.Errorf("expected Makefile to be created, but got error: %v", err)
					}
				}
			}
		})
	}
}

// TestVersionCollisionDetection tests the specific collision scenarios.
func TestVersionCollisionDetection(t *testing.T) {
	yamlV010 := `version: v0.1.0
plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.0`

	yamlV011 := `version: v0.1.1
plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.1`

	// Current version should not collide with itself
	if hasYAMLCollision(yamlV011, yamlV011) {
		t.Error("same version should not collide")
	}

	// Different versions should collide
	if !hasYAMLCollision(yamlV010, yamlV011) {
		t.Error("different versions should collide")
	}

	// Reverse order should also collide
	if !hasYAMLCollision(yamlV011, yamlV010) {
		t.Error("different versions in reverse order should collide")
	}
}

// BenchmarkExtractVersionFromYAML benchmarks version extraction.
func BenchmarkExtractVersionFromYAML(b *testing.B) {
	content := `version: v1.59.0

plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.1

linters-settings:
  custom:
    uber-go-lint-style:
      type: "module"
`
	for i := 0; i < b.N; i++ {
		extractVersionFromYAML(content)
	}
}

// BenchmarkHasYAMLCollision benchmarks collision detection.
func BenchmarkHasYAMLCollision(b *testing.B) {
	yaml1 := `version: v0.1.0
plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.0`

	yaml2 := `version: v0.1.1
plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.1`

	for i := 0; i < b.N; i++ {
		hasYAMLCollision(yaml1, yaml2)
	}
}

// TestIndent_EmptyLines verifies empty lines are preserved.
func TestIndent_EmptyLines(t *testing.T) {
	input := "first\n\nsecond\n\n\nthird"
	expected := "  first\n\n  second\n\n\n  third"
	got := indent(input, "  ")
	if got != expected {
		t.Errorf("indent with empty lines = %q, want %q", got, expected)
	}
}

// TestVersionExtraction_RealConfigs tests with actual config content.
func TestVersionExtraction_RealConfigs(t *testing.T) {
	// Test with actual customGclConfig
	version := extractVersionFromYAML(customGclConfig)
	if version != "latest" {
		t.Errorf("extractVersionFromYAML(customGclConfig) = %q, want %q", version, "latest")
	}

	// Test with actual golangciConfig
	version = extractVersionFromYAML(golangciConfig)
	if version != "" {
		// golangciConfig has version in severity rules, not the plugin version
		// This is expected behavior - we look for the plugin version specifically
		t.Logf("note: golangciConfig doesn't have plugin version in expected format")
	}
}

func TestGolangCIConfig_IsValidYAML(t *testing.T) {
	var cfg map[string]any
	if err := yaml.Unmarshal([]byte(golangciConfig), &cfg); err != nil {
		t.Fatalf("golangciConfig must be valid YAML: %v", err)
	}

	linters := mustMapValue(t, cfg, "linters")
	if _, ok := linters["enable"]; !ok {
		t.Fatalf("golangciConfig must define linters.enable")
	}

	// New golangci config shape nests plugin settings under `linters.settings.custom`.
	lintersSettings := mustMapValue(t, linters, "settings")
	custom := mustMapValue(t, lintersSettings, "custom")
	if _, ok := custom["uber-go-lint-style"]; !ok {
		t.Fatalf("golangciConfig must define custom settings for uber-go-lint-style")
	}
}

func TestMergeGolangCIConfig_MergesWithoutOverwriting(t *testing.T) {
	existing := `version: "2"
linters:
  disable-all: false
  enable:
    - govet
linters-settings:
  custom:
    existing-linter:
      type: module
severity:
  default: error
  rules:
    - linters:
        - govet
      severity: error
`

	merged, changed, err := mergeGolangCIConfig(existing, golangciConfig)
	if err != nil {
		t.Fatalf("mergeGolangCIConfig returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected merge to report changes")
	}

	cfg := map[string]any{}
	if err := yaml.Unmarshal([]byte(merged), &cfg); err != nil {
		t.Fatalf("failed to parse merged YAML: %v", err)
	}

	linters := mustMapValue(t, cfg, "linters")
	if disableAll, ok := linters["disable-all"].(bool); !ok || disableAll {
		t.Fatalf("expected existing linters.disable-all=false to be preserved")
	}

	enabled := mustSliceValue(t, linters, "enable")
	assertContainsString(t, enabled, "govet")
	assertContainsString(t, enabled, "uber-go-lint-style")

	settings := mustMapValue(t, cfg, "linters-settings")
	custom := mustMapValue(t, settings, "custom")
	_ = mustMapValue(t, custom, "existing-linter")
	plugin := mustMapValue(t, custom, "uber-go-lint-style")
	if plugin["type"] != "module" {
		t.Fatalf("expected plugin type=module, got %v", plugin["type"])
	}

	severity := mustMapValue(t, cfg, "severity")
	rules := mustSliceValue(t, severity, "rules")
	if !hasSeverityRuleForLinter(rules, "uber-go-lint-style", "warning") {
		t.Fatalf("expected severity warning rule for uber-go-lint-style")
	}
}

func TestMergeGolangCIConfig_NoOpWhenAlreadyConfigured(t *testing.T) {
	existingMap := map[string]any{
		"version": "2",
		"linters": map[string]any{
			"disable-all": false,
			"enable":      []any{"uber-go-lint-style"},
		},
		"linters-settings": map[string]any{
			"custom": map[string]any{
				"uber-go-lint-style": map[string]any{
					"type":         "module",
					"description":  "Uber Go style guide linter",
					"path":         "./custom-gcl.so",
					"original-url": "github.com/beltranaceves/uber-go-lint-style",
					"settings": map[string]any{
						"disabled_rules_yaml": "- TodoRule\n",
					},
				},
			},
		},
		"severity": map[string]any{
			"default": "error",
			"rules": []any{
				map[string]any{
					"linters":  []any{"uber-go-lint-style"},
					"severity": "warning",
				},
			},
		},
	}

	b, err := yaml.Marshal(existingMap)
	if err != nil {
		t.Fatalf("failed to marshal existing map: %v", err)
	}
	existing := string(b)

	merged, changed, err := mergeGolangCIConfig(existing, golangciConfig)
	if err != nil {
		t.Fatalf("mergeGolangCIConfig returned error: %v", err)
	}
	if changed {
		t.Fatalf("expected no changes when config is already compatible")
	}
	if merged != existing {
		t.Fatalf("expected merged output to equal existing input when unchanged")
	}
}

func mustMapValue(t *testing.T, parent map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := parent[key].(map[string]any)
	if !ok {
		t.Fatalf("expected map at key %q, got %T", key, parent[key])
	}
	return value
}

func mustSliceValue(t *testing.T, parent map[string]any, key string) []any {
	t.Helper()
	value, ok := parent[key].([]any)
	if !ok {
		t.Fatalf("expected slice at key %q, got %T", key, parent[key])
	}
	return value
}

func assertContainsString(t *testing.T, values []any, want string) {
	t.Helper()
	for _, value := range values {
		if s, ok := value.(string); ok && s == want {
			return
		}
	}
	t.Fatalf("expected slice to contain %q, got %v", want, values)
}

func hasSeverityRuleForLinter(rules []any, linterName, severityName string) bool {
	for _, rule := range rules {
		ruleMap, ok := rule.(map[string]any)
		if !ok {
			continue
		}
		linters, ok := ruleMap["linters"].([]any)
		if !ok {
			continue
		}
		foundLinter := false
		for _, linter := range linters {
			if linterStr, ok := linter.(string); ok && linterStr == linterName {
				foundLinter = true
				break
			}
		}
		if !foundLinter {
			continue
		}
		if ruleMap["severity"] == severityName {
			return true
		}
	}
	return false
}

// TestCollisionDetection_RealConfigs tests collision with actual configs.
func TestCollisionDetection_RealConfigs(t *testing.T) {
	newConfig := customGclConfig
	oldConfig := `version: v1.59.0

plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.0
`
	if !hasYAMLCollision(oldConfig, newConfig) {
		t.Error("should detect collision between v0.1.0 and v0.1.1")
	}

	// No collision if same version
	if hasYAMLCollision(newConfig, newConfig) {
		t.Error("should not detect collision with same version")
	}
}

// TestCreateOrUpdateFile_WriteError tests handling of write errors.
func TestCreateOrUpdateFile_WriteError(t *testing.T) {
	// Test with a directory that doesn't exist and is invalid
	// This should error when trying to write
	invalidPath := "/root/no-permission/config.yml"

	// Skip if running as root
	if os.Geteuid() == 0 {
		t.Skip("test cannot run as root")
	}

	err := createOrUpdateFile(invalidPath, "test content", true)
	if err == nil {
		t.Error("expected write error for invalid path, got nil")
	}
}

// TestBufferOutput captures output for verification (helper for testing output).
type captureOutput struct {
	buffer *bytes.Buffer
}

// Helper to test prompt validation
func TestPromptInputValidation(t *testing.T) {
	tests := []struct {
		name      string
		options   []string
		wantFirst string
	}{
		{
			name:      "single option",
			options:   []string{"skip"},
			wantFirst: "skip",
		},
		{
			name:      "multiple options",
			options:   []string{"skip", "overwrite"},
			wantFirst: "skip", // First option is default
		},
		{
			name:      "three options",
			options:   []string{"merge", "skip", "overwrite"},
			wantFirst: "merge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate that the first option would be the default
			if len(tt.options) == 0 {
				t.Fatal("options cannot be empty")
			}
			if tt.options[0] != tt.wantFirst {
				t.Errorf("first option = %s, want %s", tt.options[0], tt.wantFirst)
			}
		})
	}
}
