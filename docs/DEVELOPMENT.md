# Custom Revive Rules Implementation Guide

## Overview

This directory contains documentation for implementing and using custom revive rules based on Uber's Go Style Guide.

## Quick Start

1. **Explore existing rules**: Check `../rules/` directory for examples
2. **Understand revive**: Review [revive documentation](https://github.com/mgechev/revive)
3. **Use the template**: Start with `rules/example_rule.go`
4. **Configure**: Add your rule to `../revive.toml`
5. **Test**: Create tests in `rules/testdata/`

## Rule Development Workflow

### Step 1: Create Rule File

Create a new file in `rules/` directory following the naming convention: `rule_name.go`

```go
package rules

import (
    "go/ast"
    "github.com/mgechev/revive/lint"
)

type MyNewRule struct{}

func (r *MyNewRule) Name() string {
    return "my-new-rule"
}

func (r *MyNewRule) Apply(file *lint.File, arguments lint.Arguments) []lint.Failure {
    var failures []lint.Failure
    // Implementation
    return failures
}
```

### Step 2: Implement Rule Logic

Use AST traversal to find violations:

```go
func (r *MyNewRule) Apply(file *lint.File, arguments lint.Arguments) []lint.Failure {
    var failures []lint.Failure
    
    // Use a visitor pattern
    walker := &astWalker{failures: &failures}
    ast.Walk(walker, file.AST)
    
    return failures
}

type astWalker struct {
    failures *[]lint.Failure
}

func (w *astWalker) Visit(node ast.Node) ast.Visitor {
    if funcDecl, ok := node.(*ast.FuncDecl); ok {
        // Check function declarations
        // Append failures as needed
    }
    return w
}
```

### Step 3: Register in revive.toml

```toml
[[rule]]
name = "my-new-rule"
severity = "warning"
arguments = []
```

### Step 4: Test

Create test cases in `rules/testdata/my_new_rule/` with valid and invalid Go code.

## Common Rule Patterns

### Pattern 1: Check All Functions

```go
ast.Walk(&functionChecker{failures: &failures}, file.AST)
```

### Pattern 2: Check Type Declarations

```go
for _, decl := range file.AST.Decls {
    typeSpec, ok := decl.(*ast.TypeSpec)
    if !ok {
        continue
    }
    // Process type
}
```

### Pattern 3: Check Variable/Const Declarations

```go
for _, decl := range file.AST.Decls {
    genDecl, ok := decl.(*ast.GenDecl)
    if !ok || (genDecl.Tok != token.VAR && genDecl.Tok != token.CONST) {
        continue
    }
    // Process variables/constants
}
```

## Integration with Repository

1. Custom rules should be self-contained in the `rules/` directory
2. Each rule corresponds to a section in Uber's style guide
3. Configuration is managed centrally in `revive.toml`
4. Documentation should be maintained in `style_guide/rules/` for the corresponding style guide rule

## Best Practices

1. **Clarity**: Rule names should be clear and descriptive
2. **Documentation**: Include comments explaining the rule's purpose
3. **Testing**: Always write tests for new rules
4. **Performance**: Minimize AST traversals for efficiency
5. **Configurability**: Allow arguments to customize rule behavior
6. **Error Handling**: Handle edge cases gracefully

## Debugging Tips

- Print AST information: Use `ast.Print()` during development
- Enable revive verbose mode: `revive -v`
- Test isolated code: Create minimal test cases
- Check node types: Use `fmt.Printf("%T", node)` to debug

## References

- [Revive GitHub](https://github.com/mgechev/revive)
- [Uber Go Style Guide](https://github.com/uber-go/guide)
- [Go AST Documentation](https://golang.org/pkg/go/ast/)
- [Go Token Package](https://golang.org/pkg/go/token/)
