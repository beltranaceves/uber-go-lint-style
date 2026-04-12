# uber-go-lint-style

[![Go Test](https://github.com/beltranaceves/uber-go-lint-style/actions/workflows/go-test.yml/badge.svg)](https://github.com/beltranaceves/uber-go-lint-style/actions/workflows/go-test.yml)
[![Coverage Status](https://codecov.io/gh/beltranaceves/uber-go-lint-style/branch/main/graph/badge.svg)](https://codecov.io/gh/beltranaceves/uber-go-lint-style)

<p align="center">
  <img src="./assets/ACKCHYUALLY.png" alt="" width="300">
  <br>
  <!-- Logo by <a href="https://github.com/hawkgs">Georgi Serev</a> -->
</p>

A golangci-lint plugin implementing custom Go linting rules based on [Uber's Go Style Guide](https://github.com/uber-go/guide).

## Overview

This is a custom golangci-lint plugin that enforces Uber's internal Go coding standards through static analysis. It's designed to catch style violations early and guide developers toward safer, more maintainable code patterns.

## Installation

### Prerequisites

- Go 1.23+
- golangci-lint 1.59.0+

Install golangci-lint:
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Setup Option 1: Automated Setup (Recommended)

Run the setup script to auto-generate configuration files:

```bash
go run github.com/beltranaceves/uber-go-lint-style/cmd/setup@latest
```

**Note:** This requires a released version. If you want to test locally first, clone the repo and run:
```bash
go run ./cmd/setup
```

This creates:
- `.custom-gcl.yml` — Plugin configuration
- `.golangci.yml` — Linter settings
- `Makefile` — Build and run commands

Then simply:
```bash
make
```

### Setup Option 2: Manual Configuration

If you prefer manual setup, follow these steps:

**Step 1: Create `.custom-gcl.yml`**

```yaml
version: v1.59.0

plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.0  # Use latest release
```

Or for local development:
```yaml
plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    path: /path/to/uber-go-lint-style
```

**Step 2: Create a `.golangci.yml` to enable the plugin and rules**

```yaml
version: "1"

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
```

**Step 3: Build the custom binary and run**

```bash
golangci-lint custom
./custom-gcl run ./...
```

**Adding a Makefile target (optional)**

To avoid running commands manually each time, add this target to your `Makefile`:

```makefile
.PHONY: uber_lint
uber_lint:
	@if [ ! -f "./custom-gcl" ]; then \
		echo "Building custom golangci-lint with uber-go-lint-style plugin..."; \
		golangci-lint custom || exit 1; \
	fi
	@./custom-gcl run ./...
```

This target automatically builds the binary on first run and caches it for subsequent runs. Then simply:
```bash
make uber_lint
```

---

## Rules

See [RULES.md](RULES.md) for full rule descriptions and examples.

## Development

### Project Structure

```
uber-go-lint-style/
├── plugin.go                # golangci-lint plugin entry point
├── plugin_test.go           # Plugin tests
├── rules/                   # Custom rule implementations
│   ├── todo.go             # TODO rule
│   └── atomic.go           # Atomic rule
├── testdata/               # Test data for rules
├── style_guide/            # Uber style guide documentation
└── test-client/            # Client integration tests
```

### Adding a New Rule

1. Create a new file in `rules/` (e.g., `rules/myrule.go`):

```go
package rules

import (
	"golang.org/x/tools/go/analysis"
)

type MyRule struct{}

func (r *MyRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "myrule",
		Doc: "enforce your style convention",
		Run: r.run,
	}
}

func (r *MyRule) run(pass *analysis.Pass) (any, error) {
	// Your linting logic here
	return nil, nil
}
```

**Alternative approach:** For more detailed guidance on rule structure, performance patterns, and testing conventions, see `.github/skills/develop-linter-rules/SKILL.md`. This skill covers best practices, analysis approach selection, and examples.

2. Add test data in `testdata/src/testlintdata/myrule/`:

```go
package myrule_test

// Violations here
func bad() {
	// want "error message"
}

// Good practices here  
func good() {
}
```

3. Add test in `plugin_test.go`:

```go
func TestMyRule(t *testing.T) {
	// Similar to existing test patterns
}
```

4. Register in `plugin.go`:

```go
func (f *PluginExample) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		(&rules.TodoRule{}).BuildAnalyzer(),
		(&rules.AtomicRule{}).BuildAnalyzer(),
		(&rules.MyRule{}).BuildAnalyzer(),  // Add here
	}, nil
}
```

### Running Tests

```bash
go test ./...
```

## Contributing

This project implements style rules from [Uber's Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md). When adding new rules:

1. Reference the specific style guideline being enforced
2. Document how the check works in the rule's `Doc` field
3. Provide comprehensive test cases (both good and bad patterns)
4. Keep rules focused and single-purpose

## Resources

- [uber-go/guide](https://github.com/uber-go/guide) — Uber's Go style guide
- [golangci-lint plugins](https://golangci-lint.run/docs/plugins/plugins-configuration/) — Custom plugin documentation
- [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis) — Analysis framework used
- [go/ast](https://pkg.go.dev/go/ast)
- [go/types](https://pkg.go.dev/go/types)
