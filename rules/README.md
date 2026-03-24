# Custom Revive Rules

This directory contains custom linting rules for the [revive](https://github.com/mgechev/revive) linter, implementing best practices from [Uber's Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

## Directory Structure

```
rules/
├── example_rule.go        # Template for implementing custom rules
├── error_once.go          # Rule: Error variables assigned once per scope
├── error_wrap.go          # Rule: Errors properly wrapped with context
├── function_order.go      # Rule: Functions ordered per conventions
├── enum_start.go          # Rule: Enum values start from zero
├── defer_clean.go         # Rule: Defers used for cleanup
└── ... (additional rules)
```

## Implementation Guidelines

### 1. Rule Structure

Each custom rule must implement the `lint.Rule` interface from the revive package:

```go
package rules

import "github.com/mgechev/revive/lint"

type MyRule struct{}

func (r *MyRule) Name() string {
    return "my-rule-name"
}

func (r *MyRule) Apply(file *lint.File, arguments lint.Arguments) []lint.Failure {
    var failures []lint.Failure
    // Implement rule logic here
    return failures
}
```

### 2. Key Components

- **Name()**: Returns the rule identifier (used in revive.toml)
- **Apply()**: Executes the rule against the AST, returning violations
- **lint.Failure**: Represents a rule violation with location and message

### 3. Common AST Manipulation Patterns

```go
import "go/ast"

// Walk the AST to find nodes
ast.Walk(&visitor{}, file.AST)

// Check specific node types
switch node := decl.(type) {
case *ast.FuncDecl:
    // Process function declarations
case *ast.TypeSpec:
    // Process type declarations
case *ast.ValueSpec:
    // Process variable/constant declarations
}
```

### 4. Creating Failures

```go
failure := lint.Failure{
    Failure:     "Description of the violation",
    Node:        astNode,
    Confidence:  1.0, // 0.0 to 1.0
}
failures = append(failures, failure)
```

## Configuration

Rules are enabled and configured in `/revive.toml`. Example:

```toml
[[rule]]
name = "my-rule-name"
severity = "warning"
arguments = [arg1, arg2]
```

## Testing Custom Rules

To test a custom rule:

1. Create a test file in `rules/testdata/` with Go code to lint
2. Implement tests in `*_test.go` files
3. Run: `go test ./rules/...`

## Implementing Specific Uber Rules

### error-once
**Reference**: Uber Go Style Guide - Error Handling  
**Description**: Error variables should only be assigned once per scope

### error-wrap
**Reference**: Uber Go Style Guide - Error Handling  
**Description**: Errors should be wrapped with additional context

### function-order
**Reference**: Uber Go Style Guide - Function Organization  
**Description**: Functions should be ordered by visibility and responsibility

### enum-start
**Reference**: Uber Go Style Guide - Code Organization  
**Description**: Enum values should start from zero (iota)

### defer-clean
**Reference**: Uber Go Style Guide - Defer Usage  
**Description**: Defers should be reserved for cleanup operations

## Resources

- [Revive Documentation](https://github.com/mgechev/revive)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go AST Package](https://golang.org/pkg/go/ast/)
- [Go Parser Package](https://golang.org/pkg/go/parser/)
