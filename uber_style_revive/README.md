# Uber Go Style Revive

A minimal, self-contained Go linter based on [Revive](https://github.com/mgechev/revive) that enforces [Uber's Go Style Guide](https://github.com/uber-go/guide).

## Features

- **5 Core Uber Style Rules** - Atomic operations, error wrapping, error naming, struct embedding, and global mutability
- **Full Test Pipeline** - Convention-based test harness with positive/negative fixtures
- **Standalone CLI** - Run independently or as a golangci-lint plugin
- **Compatible with Revive** - Uses revive's `lint.Rule` interface and APIs
- **Minimal Dependencies** - Only depends on `github.com/mgechev/revive`

## Project Structure

```
uber_style_revive/
├── rules/               # Rule implementations
│   ├── atomic.go
│   ├── error_wrap.go
│   ├── error_name.go
│   ├── struct_embed.go
│   ├── global_mut.go
│   ├── init.go         # Rule registry
│   └── rules_test.go   # Test harness
│
├── testdata/           # Test fixtures (positive + negative cases)
│   ├── atomic/
│   ├── error-wrap/
│   ├── error-name/
│   ├── struct-embed/
│   └── global-mut/
│
├── cmd/uber-go-lint/
│   └── main.go        # CLI entry point
│
├── internal/linter/
│   └── runner.go      # Revive wrapper
│
├── go.mod
└── .golangci-uber.yml # golangci-lint configuration
```

## Rules

### 1. **atomic** - Use go.uber.org/atomic
Flags usage of `sync/atomic` instead of the type-safe `go.uber.org/atomic` package.

```go
// Bad
import "sync/atomic"
var count int64
atomic.AddInt64(&count, 1)

// Good
import "go.uber.org/atomic"
var count atomic.Int64
count.Add(1)
```

### 2. **error-wrap** - Wrap Errors with Context
Enforces wrapping errors with context using `fmt.Errorf("%w", err)`.

```go
// Bad
return err

// Good
return fmt.Errorf("operation failed: %w", err)
```

### 3. **error-name** - Standard Error Naming
Enforces error variables follow naming conventions (`err`, `parseErr`, etc).

```go
// Bad
e := json.Unmarshal(data, &obj)
if e != nil { ... }

// Good
err := json.Unmarshal(data, &obj)
if err != nil { ... }
```

### 4. **struct-embed** - Explicit Struct Embedding
Prevents embedding basic types without explicit field names.

```go
// Bad
type Config struct {
    string    // Embedded basic type
    int
}

// Good
type Config struct {
    Data    string
    Timeout int
}
```

### 5. **global-mut** - Avoid Mutable Globals
Discourages exported mutable global variables.

```go
// Bad
var Config = map[string]string{}

// Good
const DefaultTimeout = 30
var config = map[string]string{}  // Unexported
```

## Usage

### Standalone CLI

```bash
# Build the CLI
go build -o uber-go-lint ./cmd/uber-go-lint

# Lint a directory
./uber-go-lint ./...

# Lint with specific format
./uber-go-lint -format json ./...

# List available rules
./uber-go-lint -list
```

### Testing

Run the comprehensive test suite:

```bash
# All tests
go test ./rules -v

# Verbose output
go test ./rules -v -args -verbose

# Specific rule
go test ./rules -run TestAllRules/atomic
```

The test harness automatically:
- Discovers all rule directories in `testdata/`
- Loads positive fixtures (should have lint failures)
- Loads negative fixtures (should have no failures)
- Runs tests in parallel
- Reports detailed pass/fail status

### golangci-lint Integration

Use as a golangci-lint custom linter:

```bash
# Build the linter binary
go build -o bin/uber-go-lint ./cmd/uber-go-lint

# Run with golangci-lint
golangci-lint run -c .golangci-uber.yml ./...
```

Configuration in `.golangci-uber.yml`:

```yaml
linters-settings:
  custom:
    uber-go-style:
      path: ./bin/uber-go-lint
      original-url: https://github.com/beltranaceves/uber-go-lint-style
```

## Architecture

### Rule Implementation Pattern

Each rule implements the `lint.Rule` interface from revive:

```go
type MyRule struct{}

func (r *MyRule) Name() string {
    return "my-rule"
}

func (r *MyRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
    // Analyze AST and return failures
}
```

### Test Harness Design

The test harness (`rules/rules_test.go`) uses a convention-based system:

1. **Discovery** - Scans `testdata/` for rule directories
2. **Registration** - Uses `rules.NewRule()` factory to create rule instances
3. **Execution** - Runs each rule against positive/negative fixtures
4. **Assertion** - Verifies positive fixtures have ≥1 failure, negative have 0

Example test data structure:

```
testdata/atomic/
├── positive_test.go   (uses sync/atomic → should FAIL)
└── negative_test.go   (uses go.uber.org/atomic → should PASS)
```

### Adding New Rules

To add a new rule (e.g., "my-rule"):

1. Create `rules/my_rule.go` implementing `lint.Rule`
2. Create `testdata/my-rule/positive_test.go` (bad code)
3. Create `testdata/my-rule/negative_test.go` (good code)
4. Register in `rules/init.go`
5. Run tests: `go test ./rules -v`

The test harness auto-discovers it!

## Dependencies

- `github.com/mgechev/revive` - Linting framework
- Go 1.21+

## Making It Compatible with golangci-lint

This project is designed to work with golangci-lint's custom linter plugin system:

1. **Standalone CLI** - Can be run independently
2. **Revive-based** - Uses standard `lint.Rule` interface
3. **Configurable** - Rules can be turned on/off via config
4. **Pluggable** - Easily integrates into golangci-lint workflow

To use in golangci-lint:

```bash
# Build the binary
go build -o bin/uber-go-lint ./cmd/uber-go-lint

# golangci-lint runs it as a custom linter
golangci-lint run -c .golangci-uber.yml
```

## Testing Rules in Isolation

The test fixtures can be linted individually:

```bash
# Run linter on a specific test fixture
./uber-go-lint testdata/atomic/positive_test.go

# Output should show atomic-related violations
```

## Future Extensions

This minimal structure supports adding:
- More rules (just extend `init.go` and add `testdata/`)
- Custom configuration loading (TOML/YAML)
- Additional output formatters
- Integration with other linting tools
- LLM-powered rule suggestions

## References

- [Uber Go Style Guide](https://github.com/uber-go/guide)
- [Revive Linter](https://github.com/mgechev/revive)
- [golangci-lint Plugin System](https://golangci-lint.run/)
