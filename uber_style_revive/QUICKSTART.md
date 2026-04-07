# Quick Start Guide - Uber Go Style Revive

## Setup

### 1. Install Dependencies
From the `uber_style_revive` directory:

```bash
cd uber_style_revive
go mod download
```

### 2. Build the CLI Tool

```bash
# Using go build
go build -o uber-go-lint ./cmd/uber-go-lint

# Or using make
make build
```

This creates the `uber-go-lint` executable.

## Running Tests

### Run All Tests
```bash
go test ./rules -v
```

Expected output:
```
Found 5 rules to test
rule atomic               PASS
rule error-name          PASS
rule error-wrap          PASS
rule global-mut          PASS
rule struct-embed        PASS
```

### Run Tests with Verbose Output
```bash
go test ./rules -v -args -verbose
```

Shows detailed per-fixture test results.

### Run a Single Rule's Tests
```bash
go test ./rules -run TestAllRules/atomic -v
```

## Using the CLI

### Basic Linting
```bash
# Lint the current directory
./uber-go-lint ./...

# Lint specific files
./uber-go-lint main.go handler.go

# Lint with specific format
./uber-go-lint -format json ./...
```

### Available Formats
- `friendly` (default) - Human-readable with file:line:column format
- `simple` - Minimal output
- `json` - Structured JSON output

### List Available Rules
```bash
./uber-go-lint -list
```

Output:
```
Available Uber Go Style Rules:

  atomic              Enforce go.uber.org/atomic over sync/atomic
  error-name          Enforce standard error variable naming
  error-wrap          Enforce error wrapping with context
  global-mut          Discourage mutable global variables
  struct-embed        Prevent embedding of basic types without named fields
```

## Understanding the Test Structure

### Test Data Organization

```
testdata/
├── atomic/
│   ├── positive_test.go     # Bad code - uses sync/atomic directly
│   └── negative_test.go     # Good code - uses go.uber.org/atomic
├── error-name/
│   ├── positive_test.go     # Uses 'e' as error variable
│   └── negative_test.go     # Uses 'err' as error variable
└── [3 more rules...]
```

### How Tests Work

1. **Test Discovery** - `TestAllRules()` scans `testdata/` for rule directories
2. **Rule Instantiation** - Uses `rules.NewRule()` factory
3. **Fixture Execution** - Runs rule against each `.go` file
4. **Assertion**:
   - ✅ positive fixtures MUST have ≥1 lint failure
   - ✅ negative fixtures MUST have 0 lint failures
5. **Parallel Execution** - Tests run in parallel via `t.Parallel()`

### Test Harness Code

Located in `rules/rules_test.go`:

```go
func TestAllRules(t *testing.T) {
    // 1. Discover all rule directories in testdata/
    // 2. For each rule, instantiate via factory
    // 3. Run positive fixtures (expect failures)
    // 4. Run negative fixtures (expect no failures)
}
```

## Extending with New Rules

### Adding the "my-rule" Rule

#### 1. Create the Rule Implementation
Create `rules/my_rule.go`:

```go
package rules

import (
    "go/ast"
    "github.com/mgechev/revive/lint"
)

type MyRuleRule struct{}

func (r *MyRuleRule) Name() string {
    return "my-rule"
}

func (r *MyRuleRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
    var failures []lint.Failure
    // Analyze AST and detect violations
    return failures
}
```

#### 2. Create Test Data
Create `testdata/my-rule/positive_test.go` (bad code):

```go
package testdata

// Example of bad code that violates my-rule
```

Create `testdata/my-rule/negative_test.go` (good code):

```go
package testdata

// Example of good code that complies with my-rule
```

#### 3. Register in init.go

Update `rules/init.go`:

```go
func NewRule(name string) (lint.Rule, bool) {
    factories := map[string]func() lint.Rule{
        "atomic":       func() lint.Rule { return &AtomicRule{} },
        "error-wrap":   func() lint.Rule { return &ErrorWrapRule{} },
        // ... existing rules ...
        "my-rule":      func() lint.Rule { return &MyRuleRule{} },  // Add this
    }
    // ...
}

func GetAllRuleNames() []string {
    return []string{
        "atomic",
        "error-wrap",
        // ... existing rules ...
        "my-rule",  // Add this
    }
}
```

#### 4. Run Tests
```bash
go test ./rules -v
```

The test harness auto-discovers your new rule!

## Integration with golangci-lint

### Build and Configure

```bash
# Build the linter binary
go build -o bin/uber-go-lint ./cmd/uber-go-lint

# Run with golangci-lint (uses .golangci-uber.yml)
golangci-lint run -c .golangci-uber.yml ./...
```

### Configuration File

`.golangci-uber.yml` already contains:

```yaml
linters-settings:
  custom:
    uber-go-style:
      path: ./bin/uber-go-lint
      description: Uber Go Style Guide Linter
```

## Development Tips

### Debug a Rule
To see what a rule catches, lint against the positive fixture:

```bash
./uber-go-lint testdata/atomic/positive_test.go
```

Should show violations.

### Verify Rule Works Correctly
Lint against the negative fixture:

```bash
./uber-go-lint testdata/atomic/negative_test.go
```

Should show no violations (exit code 0).

### Check Test Output
See detailed test results:

```bash
go test ./rules -v -run atomic
```

### Modify and Re-test
After changing a rule:

```bash
# Rebuild CLI
go build -o uber-go-lint ./cmd/uber-go-lint

# Run specific rule's tests
go test ./rules -v -run TestAllRules/atomic
```

## Project Architecture

```
uber_style_revive/
├── rules/                 ← Rule implementations + test harness
│   ├── atomic.go          ← Individual rule
│   ├── init.go            ← Registry/factory
│   └── rules_test.go      ← Test harness (auto-discovers rules)
│
├── testdata/              ← Convention-based fixtures
│   ├── atomic/
│   │   ├── positive_test.go
│   │   └── negative_test.go
│   └── [4 more rules...]
│
├── cmd/uber-go-lint/      ← CLI entry point
│   └── main.go
│
├── internal/linter/       ← Revive wrapper
│   └── runner.go
│
└── go.mod                 ← Dependencies
```

## Troubleshooting

### Tests Fail to Discover Rules
✅ Ensure directories in `testdata/` match rule names exactly

### Rule Not Running
✅ Verify it's registered in `rules/init.go`

### CLI Can't Find Rules
✅ Build with: `go build -o uber-go-lint ./cmd/uber-go-lint`

### golangci-lint Can't Execute
✅ Build binary: `go build -o bin/uber-go-lint ./cmd/uber-go-lint`

## Key Concepts

### Convention-Based Testing
- Rule directories auto-discovered by name
- No manual rule registration in test harness
- Positive/negative fixtures follow naming convention
- Test harness is generic and rule-agnostic

### Revive Integration
- Rules implement `lint.Rule` interface
- Uses revive's AST analysis capabilities
- Compatible with revive's linting pipeline
- Can be used as revive plugin

### Standalone + Plugin Architecture
- Runs as independent CLI tool
- Works as golangci-lint custom linter
- No lock-in to any single framework
- Minimal dependencies
