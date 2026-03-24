# Uber Go Linter Custom Rules Development

This directory contains the source code for custom revive linter rules based on Uber's Go style guide.

## Project Structure

- **`.golangci.yml`** - Main linter configuration file that orchestrates all linters
- **`revive.toml`** - Revive-specific configuration enabling standard rules and custom rules
- **`rules/`** - Custom revive rule implementations in Go
  - `example_rule.go` - Template for implementing new custom rules
  - (Add your custom rules here)

## Linting Setup

This project follows Uber's recommended linting setup from the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md#linting):

### Enabled Linters

1. **errcheck** - Ensures error handling is not missed
2. **goimports** - Manages import formatting and organization
3. **revive** - Enforces style mistakes (modern replacement for deprecated `golint`)
4. **govet** - Catches common Go mistakes
5. **staticcheck** - Performs static analysis checks

### Using golangci-lint

Install golangci-lint:
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

Run linters:
```bash
golangci-lint run ./...
```

Run specific linter:
```bash
golangci-lint run --linters=revive ./...
```

## Developing Custom Revive Rules

### Understanding Revive Rules

Revive rules must implement the `lint.Rule` interface with:
- `Name()` - Returns the rule identifier
- Apply() - Performs the linting logic on an AST

### Creating a New Rule

1. Create a new Go file in `rules/` (e.g., `my_custom_rule.go`)
2. Implement the `lint.Rule` interface
3. Add the rule to `revive.toml` with appropriate configuration
4. Test against code samples

### Example Rule Structure

```go
package rules

import (
    "github.com/mgechev/revive/lint"
)

type MyCustomRule struct{}

func (r *MyCustomRule) Name() string {
    return "my-custom-rule"
}

func (r *MyCustomRule) Apply(file *lint.File, arguments lint.Arguments) []lint.Failure {
    var failures []lint.Failure
    // Implementation here
    return failures
}
```

### Resources

- [Revive Documentation](https://github.com/mgechev/revive)
- [Revive Rule Examples](https://github.com/mgechev/revive/tree/master/rules)
- [Go AST Package](https://pkg.go.dev/go/ast)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

## Configuration Files

### `.golangci.yml`

Main configuration orchestrating all linters. Specifies:
- Which linters to enable
- Timeout and skip patterns
- Per-linter settings

### `revive.toml`

Revive-specific configuration with:
- Individual rule configurations
- Severity levels
- Rule-specific arguments

## Next Steps

1. Review `rules/example_rule.go` for the basic rule template
2. Create custom rules based on Uber's style guide principles
3. Add rules to `revive.toml` to enable them
4. Test with `golangci-lint run ./...`
5. Iterate on rule implementations based on feedback

## Testing Custom Rules

To test your rules on sample code:

```bash
# Create a test file with violations
cat > test_sample.go << 'EOF'
package main

// Your test code here
func main() {
}
EOF

# Run linters
golangci-lint run test_sample.go

# Clean up
rm test_sample.go
```
