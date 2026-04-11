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

## Choosing Your Analysis Approach

Before implementing a rule, you need to decide which analysis packages and techniques to use. The choice depends on what information you need to inspect code.

### The Three Approaches

| Approach | Packages | Use Case | Speed | Complexity |
|----------|----------|----------|-------|------------|
| **AST-only** | `go/ast` | Structural checks: naming, comments, syntax | ⚡ Fast | Low |
| **Type-aware** | `go/ast` + `go/types` | Semantic checks: types, functions, packages | 🐢 Slower | Medium |
| **Full Analysis Framework** | `golang.org/x/tools/go/analysis` | Production-grade rules with caching and reporting | 📦 Optimized | High |

### Quick Decision Tree

**Answer these questions to know what you need:**

1. **Are you checking structure or semantics?**
   - Structure (naming, syntax): AST-only
   - Semantics (types, function signatures): Type-aware or full analysis

2. **Do you need to understand types and resolve function calls?**
   - No: AST-only (`go/ast`)
   - Yes: Type-aware (`go/ast` + `go/types`) or full analysis (`go/analysis`)

3. **Is this a production linter rule?**
   - No: Use `go/ast` or `go/types` directly
   - Yes: Use `go/analysis` framework (this project's approach)

### When to Use Each

#### **AST-Only (go/ast)**

Use when you:
- Check variable/function names or naming conventions
- Detect specific comment patterns
- Validate syntax structure (e.g., "must have a receiver")
- Look for specific code patterns without understanding their types

**Example:** `todo` rule checks for "// TODO:" comments without caring what code they annotate.

**Code pattern:**
```go
ast.Inspect(file, func(n ast.Node) bool {
	if comment, ok := n.(*ast.Comment); ok {
		// Check the comment text
	}
	return true
})
```

#### **Type-Aware (go/ast + go/types)**

Use when you:
- Need to resolve function signatures or method receivers
- Check if a variable is a specific type (e.g., `*sync.Mutex`)
- Validate that code uses a particular package (e.g., `sync/atomic`)
- Understand relationships between types (e.g., "does this implement an interface?")

**Example:** `atomic` rule checks if a function call is to `sync/atomic` using type information to resolve which package the function comes from.

**Code pattern:**
```go
// Requires pass.TypesInfo from analysis.Analyzer
if callObj, ok := pass.TypesInfo.Uses[selector]; ok {
	if fn, ok := callObj.(*types.Func); ok {
		if fn.Pkg().Path() == "sync/atomic" {
			// Found a sync/atomic call
		}
	}
}
```

#### **Full Analysis Framework (go/analysis)**

Use when you:
- Need a production-grade linter rule (this project's use case)
- Want integrated reporting, caching, and runner support
- Plan integration with golangci-lint or other tools
- Need sophisticated dependency analysis (rule A depends on rule B)

**Advantages:**
- Built-in error handling and diagnostic formatting
- Automatic integration with golangci-lint
- Efficient caching and parallel execution
- Cleaner API for common patterns

**Why this project uses it:** uber-go-lint-style is a plugin that composes multiple rules with a shared framework. `go/analysis` provides the infrastructure for registering, running, and reporting diagnostics.

### Interactive Guide

If you're unsure which approach your new rule needs, answer these questions:

**Question 1: What are you checking?**
- Naming or structure → AST-only
- Types or package membership → Ask Question 2
- Complex semantics → Ask Question 2

**Question 2: Are you building a standalone linter or a plugin?**
- Standalone/learning → `go/ast` + `go/types`
- Plugin for golangci-lint or uber-go-lint-style → Full `go/analysis`

**Question 3: Do you have existing test infrastructure for this rule?**
- No test framework → Start with `go/ast` or `go/types`
- Yes, golangci-lint plugin framework → Use `go/analysis` (you're in the right place)

### This Project's Approach

**uber-go-lint-style uses the full `go/analysis` framework** because:
1. It's a golangci-lint plugin (requires `go/analysis`)
2. Multiple rules share infrastructure (plugin lifecycle, test framework)
3. Rules need type information (`LoadModeTypesInfo`)
4. It's production code with test coverage requirements

When you create a rule here, you're always using:
- `golang.org/x/tools/go/analysis` for the framework
- `go/ast` for AST inspection
- `go/types` for type information (if your rule needs it)

The framework handles registration, running, caching, and reporting.

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

### 5. Diagnostic Severity: All Diagnostics Are Errors

**Important:** The `analysis.Diagnostic` struct has no severity level. All diagnostics reported via `pass.Report()` are treated as errors by golangci-lint — there is no native way to distinguish "warnings" from "errors."

**What this means:**
- Every `pass.Report()` call creates a lint violation that must be fixed or suppressed
- You cannot create optional "suggestions" or "warnings" through severity levels
- The framework treats all diagnostics equally

**If your rule feels more advisory than mandatory:**

1. **Document the rule's intent clearly** — Explain in the `Doc` field that this is a stylistic suggestion
2. **Document the `nolint` directive** — Tell users they can suppress it where it doesn't apply
3. **Use clear messaging** — Phrase diagnostics as suggestions ("consider using X instead of Y")

**Example:**

```go
func (r *MyNewRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "mynewrule",
		Doc: `suggest using pattern X instead of pattern Y.

			This rule detects pattern Y and recommends X as a better alternative
			for clarity and consistency. If you need to use pattern Y, you can
			suppress this rule with the //nolint:mynewrule directive.`,
		Run: r.run,
	}
}
```

### 6. Suppressing Rules with //nolint Directives

Users can suppress your rule using the standard Go `//nolint` comment, which golangci-lint recognizes:

**Suppress all linters on the next line:**

```go
//nolint
func FunctionThatTriggersMyRule() {
}
```

**Suppress a specific rule by name:**

```go
//nolint:mynewrule
func FunctionThatTriggersMyRule() {
}
```

**Suppress multiple specific rules:**

```go
//nolint:mynewrule,atomic
func FunctionThatMightTriggerMultipleRules() {
}
```

**Suppress inline (for inline violations):**

```go
func BadExample(error string) {  //nolint:mynewrule
	// ...
}
```

**Best practice:** Document this in your rule's `Doc` field so users know they have an escape hatch for edge cases where your rule's recommendation doesn't apply.

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

### Pre-Implementation: Define Your Rule's Scope

Before writing code, clarify what your rule needs:

**Ask yourself:**
1. **What code pattern are you checking?** (syntax, types, package usage, etc.)
2. **Do you need type information?** (If you're checking `*sync.Mutex` or function signatures, yes)
3. **Is your check purely syntactic?** (If you're checking naming or detecting comments, likely no)

**Rule of thumb:**
- If you don't know your answers, ask Copilot for guidance using this interaction:
  - "I'm creating a rule that checks [your pattern]. Do I need type information?"
  - Copilot can help you determine if you need `LoadModeTypesInfo` or just AST inspection

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
