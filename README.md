# uber-go-lint-style

A golangci-lint plugin implementing custom Go linting rules based on [Uber's Go Style Guide](https://github.com/uber-go/guide).

## Overview

This is a custom golangci-lint plugin that enforces Uber's internal Go coding standards through static analysis. It's designed to catch style violations early and guide developers toward safer, more maintainable code patterns.

## Features

- **`todo` rule** — Detects TODO comments without an author attribution
- **`atomic` rule** — Detects usage of `sync/atomic` on raw types; enforces `go.uber.org/atomic` for type safety
- **Extensible** — Easy to add new rules following the patterns established

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

Or if you created a Makefile (via the setup script):
```bash
make
```

---

**💡 Tip:** Use the automated setup from Option 1 — it generates all these files for you automatically!

## Rules

### `todo` — Require author in TODO comments

**What it detects:**
```go
// TODO: fix this  // ❌ VIOLATION - no author
// TODO(): fix this // ❌ VIOLATION - malformed

// TODO(alice): fix this  // ✅ OK - has author
```

**Why:** Unattributed TODOs can be lost or unmaintained. Requiring an author ensures accountability and provides context for future developers.

### `atomic` — Use go.uber.org/atomic for raw types

**What it detects:**
```go
var counter int32
atomic.StoreInt32(&counter, 1)  // ❌ VIOLATION - raw type

val := atomic.LoadInt32(&counter)  // ❌ VIOLATION - returns raw type
```

**Correct usage:**
```go
counter := atomic.NewInt32(0)
counter.Store(1)  // ✅ OK - type-safe wrapper
val := counter.Load()
```

**Why:** The `sync/atomic` package operates on raw types, making it easy to forget atomic operations. `go.uber.org/atomic` provides type-safe wrappers that prevent accidental non-atomic access.

**How the check works:**
The rule inspects the function signature of `sync/atomic` calls and flags those that take or return raw types (int32, int64, uint32, uint64, uintptr). These should be replaced with equivalent operations from `go.uber.org/atomic`.

## Testing Locally

A test-client project is included to validate the plugin:

```bash
cd test-client
make
```

This will:
1. Build the custom golangci-lint binary with the plugin
2. Run the linter against sample code with intentional violations

See [test-client/README.md](test-client/README.md) for details.

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
- [golang.org/x/tools/go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis) — Analysis framework used
