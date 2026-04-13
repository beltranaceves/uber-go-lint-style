# uber-go-lint-style

[![Go Test](https://github.com/beltranaceves/uber-go-lint-style/actions/workflows/go-test.yml/badge.svg)](https://github.com/beltranaceves/uber-go-lint-style/actions/workflows/go-test.yml)
[![Coverage Status](https://codecov.io/gh/beltranaceves/uber-go-lint-style/branch/main/graph/badge.svg)](https://codecov.io/gh/beltranaceves/uber-go-lint-style)
[![Go Report Card](https://goreportcard.com/badge/github.com/beltranaceves/uber-go-lint-style)](https://goreportcard.com/report/github.com/beltranaceves/uber-go-lint-style)

A golangci-lint plugin for [Uber's Go Style Guide](https://github.com/uber-go/guide).

<p align="center">
  <img src="./assets/ACKCHYUALLY.png" alt="" width="300">
  <br>
  <!-- Logo by <a href="https://">origin</a> -->
</p>

> [!CAUTION]
> **Disclaimer**: this project contains significant amounts of auto-generated code, pending *thorough review*.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
	- [Prerequisites](#prerequisites)
	- [Setup Option 1: Automated Setup (Recommended)](#setup-option-1-automated-setup-recommended)
	- [Setup Option 2: Manual Configuration](#setup-option-2-manual-configuration)
- [Rules](#rules)
- [Development](#development)
	- [Project Structure](#project-structure)
	- [Adding a New Rule](#adding-a-new-rule)
	- [Running Tests](#running-tests)
- [Contributing](#contributing)
- [Resources](#resources)

## Overview

This is a custom linter that strives to enforce Uber's internal Go coding standards through static analysis. It's designed to catch style violations early and guide developers toward safer, more maintainable code patterns.

## Installation

### Prerequisites

- Go 1.23+
- golangci-lint 1.59.0+ ([Install docs](https://golangci-lint.run/usage/install/))


> [!TIP]
> If you are using a coding Agent (Claude Code, AmpCode, Cursor, Copilot, etc.), copy and paste this prompt:
> ```bash
> Fetch the install guide and follow it:
> curl -s https://raw.githubusercontent.com/beltranaceves/uber-go-lint-style/refs/heads/main/installation.md
> ```

Follow these steps:

### Setup Option 1: Automated Setup (Recommended)

Run the setup script to auto-generate configuration files:

```bash
go run github.com/beltranaceves/uber-go-lint-style/cmd/setup@latest
```

This creates:
- `.custom-gcl.yml` — Plugin configuration
- `.golangci.yml` — Linter settings
- `Makefile` — Build and run commands

Then simply:
```bash
make uber_lint
```

> [!NOTE]  
> This option requires a released version. If you want to test locally first, clone the repo and run:
> ```bash
> go run ./cmd/setup
> ```

### Setup Option 2: Manual Configuration

If you prefer manual setup, follow these steps:

**Step 1: Create `.custom-gcl.yml`**

```yaml
version: v1.59.0

plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.1  # Use latest release
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

severity:
  default-severity: error
  rules:
    - linters:
        - uber-go-lint-style
      severity: warning
```

**Step 3: Build the custom binary and run**

```bash
golangci-lint custom
./custom-gcl run ./...
```

**Step 4: Add a Makefile (optional)**

To avoid running commands manually each time, add these targets to your `Makefile`:

```makefile
.DEFAULT_GOAL := uber_lint

# Run linter (builds plugin if needed)
uber_lint:
	@if [ ! -f "./custom-gcl" ]; then \
		echo "Building custom golangci-lint with uber-go-lint-style plugin..."; \
		golangci-lint custom || exit 1; \
	fi
	@./custom-gcl run
```

This automatically builds the binary on first run and caches it for subsequent runs. Then simply:
```bash
make uber_lint
```

Optional targets: `make uber_help` for usage, `make uber_clean` to reset.

```makefile

# View help
uber_help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  make uber_lint       Build plugin (if needed) and run linter"
	@echo "  make uber_clean      Remove cached plugin binary"
	@echo ""
	@echo "Examples:"
	@echo "  make uber_lint       # First run builds plugin, subsequent runs are fast"
	@echo "  make uber_clean      # Reset and rebuild plugin next time"

.PHONY: uber_lint uber_help uber_clean
uber_clean:
	@rm -f custom-gcl*
	@echo "Cleaned custom linter artifacts"
```

---

## Rules

See [RULES.md](RULES.md) for full rule descriptions and examples.

## Development

### Project Structure

```
uber-go-lint-style/
├── plugin.go                # golangci-lint plugin entry point
├── plugin_test.go           # plugin tests
├── rules/                   # rule implementations (one file per rule)
├── testdata/                # testdata used by rule tests
├── cmd/                     # helper CLI tools (e.g., setup)
│   └── setup/               # setup command source
├── style_guide/             # generated and source docs for the style guide
│   └── rules/               # markdown source files for the guide
├── test-client/             # integration test client and examples
├── assets/                  # images and other assets
├── Makefile                 # convenience targets
├── installation.md          # installation instructions
└── RULES.md                 # rule descriptions and examples
```

### Adding a New Rule

> [!NOTE]
> If you are using coding Agents, or looking for more detailed guidance on rule structure, performance patterns, and testing conventions, there are two included [skills](.github/skills/):
> - `.github/skills/develop-linter-rules/SKILL.md` covers rule structure, analysis approaches, performance considerations, and examples.
> - `.github/skills/create-linter-tests/SKILL.md` helps scaffold test cases and edge-case coverage to reduce boilerplate.

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

2. Add test data in `testdata/src/testlintdata/myrule/`:

```go
package myrule_test

// Violations here
func bad() {
	undesirable code // want "error message"
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
- Analysis tools:
  - [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)
  - [go/ast](https://pkg.go.dev/go/ast)
  - [go/types](https://pkg.go.dev/go/types)
