---
name: "develop-linter-rules"
description: "Learn how to create maintainable linter rules for the uber-go-lint-style golangci-lint plugin, following framework conventions, testing patterns, and performance best practices."
---

# Developing New Linter Rules for uber-go-lint-style

This skill teaches how to create new linting rules compatible with the `uber-go-lint-style` golangci-lint plugin framework. The framework is built on Go's `golang.org/x/tools/go/analysis` package and follows conventions for maintainability, testability, and performance.

## Framework Architecture

### Overview

`uber-go-lint-style` is a golangci-lint plugin that enforces Uber's Go coding standards through static analysis. The framework consists of:

- **Plugin Entry Point** (`plugin.go`): Registers analyzers and manages plugin lifecycle
- **Rule Implementations** (`rules/`): Individual rule logic using the `analysis.Analyzer` interface
- **Test Data** (`testdata/src/testlintdata/`): Structured test cases for each rule
- **Test Suite** (`plugin_test.go`): Framework for running analysis tests

### Key Files and Responsibilities

| File | Purpose |
|------|---------|
| `plugin.go` | Registers the plugin, manages load mode (Syntax vs TypesInfo), exposes all analyzers |
| `rules/*.go` | Implements individual linting rules |
| `testdata/src/testlintdata/<rulename>/` | Test files with violations and expected diagnostics |
| `plugin_test.go` | Test runners that execute analyzers against test data |

## Rule Implementation Pattern

Every rule follows a consistent structure:

```go
package rules

import (
	"golang.org/x/tools/go/analysis"
	// ... other imports based on rule requirements
)

// RuleNameRule implements a specific style convention.
type RuleNameRule struct{}

// BuildAnalyzer returns the analysis.Analyzer for this rule.
// This method MUST exist on every rule type.
func (r *RuleNameRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "rule_name",  // Unique identifier (snake_case)
		Doc:  `Clear explanation of what the rule checks.

			Include:
			- What code patterns this detects
			- Why the pattern is problematic
			- How to identify if your rule needs TypesInfo vs Syntax
			- Expected false positives or edge cases`,
		Run: r.run,  // Pointer to the analyzer function
	}
}

// run performs the actual linting logic.
func (r *RuleNameRule) run(pass *analysis.Pass) (any, error) {
	// Walk the AST or use type information to detect violations
	// Report diagnostics for each violation
	return nil, nil
}
```

## The Doc Field: Essential Documentation

The `Doc` field in `analysis.Analyzer` is critical. It should clearly explain:

1. **What the rule detects** — Describe the specific code pattern or convention being checked
2. **Why it matters** — Reference the style guide or explain the motivation
3. **Technical details** — Mention whether you rely on AST-only inspection or type information
4. **Known limitations** — Document edge cases or patterns you intentionally don't check

Example from `builtin-name` rule:

```go
Doc: `avoid using predeclared identifiers for variable and field names.

	Go has several predeclared identifiers (types, constants, functions).
	Reusing these names as variable or field names can shadow the original within
	the current lexical scope and make code confusing or hard to grep.
	
	This rule dynamically retrieves the list of predeclared identifiers from
	go/types.Universe, ensuring compatibility with all Go versions. It checks
	variable declarations, function parameters, receiver parameters, and struct
	fields for any shadowing of these built-in names.`,
```

## Core Patterns

### Pattern 1: AST Walking (Syntax-Only)

Use when you only need to examine code structure, not types. This is faster and requires less information.

```go
func (r *TodoRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Type assert to the AST node type you care about
			if comment, ok := n.(*ast.Comment); ok {
				if strings.HasPrefix(comment.Text, "// TODO:") {
					pass.Report(analysis.Diagnostic{
						Pos:     comment.Pos(),
						Message: "TODO comment has no author",
					})
				}
			}
			return true  // Continue walking
		})
	}
	return nil, nil
}
```

**When to use:**
- Checking naming conventions
- Detecting specific comment patterns
- Validating syntax structure without type resolution
- Lightweight checks that don't require package context

### Pattern 2: Type Analysis (TypesInfo)

Use when you need to resolve types or understand function signatures and package context.

```go
func (r *AtomicRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if callExpr, ok := n.(*ast.CallExpr); ok {
				if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					// Use pass.TypesInfo to resolve the actual function being called
					if callObj, ok := pass.TypesInfo.Uses[selectorExpr.Sel]; ok {
						if fn, ok := callObj.(*types.Func); ok {
							// Now you can inspect the function signature
							if fn.Pkg().Path() == "sync/atomic" {
								// Check the signature
								pass.Report(analysis.Diagnostic{
									Pos:     callExpr.Pos(),
									Message: "use go.uber.org/atomic instead",
								})
							}
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
```

**When to use:**
- Checking type signatures or return types
- Validating that code uses specific packages
- Ensuring type-safe patterns
- Any check that requires knowing whether a value is a raw type vs wrapped type

### Pattern 3: Helper Functions for Reusable Logic

Extract logic into helper functions to keep the `run` method clean and promote reusability.

**Good approach:**

```go
// isRawType checks if a type is one that should use go.uber.org/atomic.
// Dynamically checks type characteristics rather than maintaining a hardcoded list.
func isRawType(t types.Type) bool {
	if t == nil {
		return false
	}

	switch u := t.Underlying().(type) {
	case *types.Basic:
		kind := u.Kind()
		// Check for specific atomic-unsafe types
		return kind == types.Int32 || kind == types.Int64 ||
			kind == types.Uint32 || kind == types.Uint64 ||
			kind == types.Uintptr
	case *types.Pointer:
		// Recursively check pointer targets
		if basic, ok := u.Elem().Underlying().(*types.Basic); ok {
			kind := basic.Kind()
			return kind == types.Int32 || kind == types.Int64 ||
				kind == types.Uint32 || kind == types.Uint64 ||
				kind == types.Uintptr
		}
	}
	return false
}

func (r *AtomicRule) run(pass *analysis.Pass) (any, error) {
	// ... simplified with helper
	if functionInvolvesRawType(fn) {  // Defined elsewhere
		pass.Report(...)
	}
	return nil, nil
}
```

## Best Practices

### 1. Avoid Hardcoded Lists — Use Dynamic Lookups

**Avoid this:**

```go
var builtinNames = []string{"error", "string", "int", "int32", ...}  // Outdated!
```

**Do this instead:**

```go
var builtinNames = func() map[string]bool {
	names := make(map[string]bool)
	for _, name := range types.Universe.Names() {  // Dynamically computed
		names[name] = true
	}
	return names
}()
```

**Why:**
- Future-proof: Compatible with new Go versions automatically
- Maintainable: No risk of missing new builtins
- Correct: Uses the authoritative source (types.Universe)

### 2. Determine Your LoadMode

In `plugin.go`, `GetLoadMode()` tells golangci-lint what information your rules need:

```go
func (f *PluginExample) GetLoadMode() string {
	// EITHER:
	return register.LoadModeSyntax      // Fast! AST only, no type info
	// OR:
	return register.LoadModeTypesInfo   // Slower, but provides type hints
}
```

**LoadModeSyntax** (`register.LoadModeSyntax`):
- Fast, low memory overhead
- Only AST inspection available
- Best for: naming, comment, structural checks
- Examples: `todo` rule (just checks comment text)

**LoadModeTypesInfo** (`register.LoadModeTypesInfo`):
- Requires type checking, slower
- Full `pass.TypesInfo` available
- Best for: type-based checks, package validation
- Examples: `atomic` rule (checks function signatures)

**Current setting:** The plugin uses `LoadModeTypesInfo` because multiple rules need type information.

### 3. Performance Considerations

- **Benchmark before optimizing**: Only add complexity after measuring
- **Use appropriate algorithms**: For small sets (builtin names), maps are fine; for large sets, consider tries or other data structures
- **Avoid redundant checks**: Combine multiple checks into one AST traversal when possible
- **Cache computed values**: Like `builtinNames`, compute once at init time

**Example from the codebase:**

```go
// Computed once, reused for all files/checks
var builtinNames = func() map[string]bool {
	names := make(map[string]bool)
	for _, name := range types.Universe.Names() {
		names[name] = true
	}
	return names
}()
```

### 4. Clear and Actionable Diagnostics

Your diagnostic messages should help developers fix issues:

```go
// Good: Specific and actionable
pass.Report(analysis.Diagnostic{
	Pos:     name.Pos(),
	Message: "identifier 'error' shadows a built-in, consider using a different name",
})

// Bad: Vague
pass.Report(analysis.Diagnostic{
	Pos:     name.Pos(),
	Message: "bad name",  // Doesn't explain the problem
})
```

## Testing Conventions

### Test File Organization

Tests are organized in `testdata/src/testlintdata/<rulename>/`:

```
testdata/
└── src/
    └── testlintdata/
        ├── todo/
        │   └── todo.go           # Test cases with violations
        ├── atomic/
        │   └── atomic.go
        └── builtin_name/
            └── builtin_name.go
```

### Test File Format

Test files use special comments to mark expected violations:

```go
package example

// BAD: Violation detected
func example1(error string) {  // want "identifier 'error' shadows a built-in, consider using a different name"
	// `error` shadows the builtin
}

// GOOD: No violation
func example2(msg string) {
	// Correct naming
}
```

**Key points:**
- `// want "exact message"` comments mark where you expect a diagnostic
- The message must match your rule's diagnostic message exactly
- Multiple violations on one line: use multiple `// want` comments
- No `// want` = analyzer should not report a diagnostic for that line

### Test Implementation

In `plugin_test.go`, use the `analysistest` package:

```go
func TestBuiltinNameRule(t *testing.T) {
	// 1. Get the plugin
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	// 2. Instantiate it
	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	// 3. Build all analyzers
	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// 4. Run the specific analyzer against test data
	// analyzers[2] is the BuiltinNameRule (in plugin.go order)
	analysistest.Run(t, testdataDir(t), analyzers[2], "testlintdata/builtin_name")
}

func testdataDir(t *testing.T) string {
	t.Helper()
	_, testFilename, _, ok := runtime.Caller(1)
	if !ok {
		require.Fail(t, "unable to get current test filename")
	}
	return filepath.Join(filepath.Dir(testFilename), "testdata")
}
```

**Critical**: The order of analyzers in `plugin.go` `BuildAnalyzers()` matters — the test uses index position.

## Step-by-Step: Adding a New Rule

### Step 1: Create the Rule Implementation

File: `rules/mynewrule.go`

```go
package rules

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// MyNewRule enforces a specific style convention.
type MyNewRule struct{}

func (r *MyNewRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "mynewrule",
		Doc: `enforce my new style convention.

			This rule checks for [specific pattern].
			Here's why it matters: [explanation].
			
			It uses AST inspection only, so no type information is needed.`,
		Run: r.run,
	}
}

func (r *MyNewRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Your linting logic here
			return true
		})
	}
	return nil, nil
}
```

### Step 2: Create Test Data

File: `testdata/src/testlintdata/mynewrule/mynewrule.go`

```go
package mynewrule

// BAD: Violates the rule
func bad() {
	// want "your diagnostic message"
}

// GOOD: Follows the rule
func good() {
}
```

### Step 3: Register in plugin.go

Add your rule to the `BuildAnalyzers()` method:

```go
func (f *PluginExample) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		(&rules.TodoRule{}).BuildAnalyzer(),
		(&rules.AtomicRule{}).BuildAnalyzer(),
		(&rules.BuiltinNameRule{}).BuildAnalyzer(),
		(&rules.MyNewRule{}).BuildAnalyzer(),  // ADD HERE
	}, nil
}
```

**Note:** Order matters for tests — each test uses the index position.

### Step 4: Add Test in plugin_test.go

```go
func TestMyNewRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// Use the index of your rule (0=todo, 1=atomic, 2=builtin_name, 3=mynewrule)
	analysistest.Run(t, testdataDir(t), analyzers[3], "testlintdata/mynewrule")
}
```

### Step 5: Ensure Correct LoadMode

In `plugin.go`, verify `GetLoadMode()` is appropriate:

```go
func (f *PluginExample) GetLoadMode() string {
	// If your rule only needs AST: LoadModeSyntax
	// If you need types: LoadModeTypesInfo (current setting)
	return register.LoadModeTypesInfo
}
```

If **all** rules only need syntax, consider changing to `LoadModeSyntax` for better performance. Otherwise, stay with `LoadModeTypesInfo`.

### Step 6: Run Tests

```bash
go test ./...
```

### Step 7: Test with Multiple Violations

Enhance your test data to validate edge cases:

```go
// Multiple violations on one line (use multiple comments)
func bad(error string, int int) {  // want "identifier 'error' shadows..."
	// want "identifier 'int' shadows..."
}
```

## Debugging Tips

### Enable AST Printing

View the AST structure for test patterns:

```bash
go run golang.org/x/tools/cmd/astdump@latest ./testdata/src/testlintdata/mynewrule/mynewrule.go
```

### Inspect TypesInfo

Add debug prints in your `run` method:

```go
if fn, ok := callObj.(*types.Func); ok {
	fmt.Printf("DEBUG: Function %s from package %s\n", fn.Name(), fn.Pkg().Path())
	sig := fn.Type().(*types.Signature)
	fmt.Printf("  Params: %d, Results: %d\n", sig.Params().Len(), sig.Results().Len())
}
```

### Check Test Diagnostic Matching

Ensure your `// want` comments match exactly. The `analysistest` package does string matching:

```go
// This will FAIL if your message is:
// want "bad name"  <- Case matters, message must be exact

// Must match what you're reporting:
pass.Report(analysis.Diagnostic{
	Message: "identifier 'error' shadows a built-in, consider using a different name",
})
```

## Related Files and References

- **uber-go/guide**: The authoritative style documentation — reference specific rules when adding new checks
- **golang.org/x/tools/go/analysis**: The analysis framework documentation
- **golangci-lint plugin docs**: https://golangci-lint.run/docs/plugins/

## Summary

Creating a new rule:

1. ✅ Implement `RuleType` with `BuildAnalyzer()` and `run()` methods
2. ✅ Write clear, actionable diagnostic messages
3. ✅ Create test data with `// want` comments for expected violations
4. ✅ Register in `plugin.go`
5. ✅ Add test function in `plugin_test.go`
6. ✅ Choose appropriate load mode (Syntax vs TypesInfo)
7. ✅ Avoid hardcoded lists; use dynamic lookups where possible
8. ✅ Document your rule's behavior in the `Doc` field

The framework is designed for maintainability — keep rules focused, document assumptions, and prefer clarity over cleverness.
