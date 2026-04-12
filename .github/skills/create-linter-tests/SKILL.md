---
name: "create-linter-tests"
description: "Automate and standardize creation of test cases for a single linter rule to increase coverage and catch edge cases."
---

# Creating Linter Test Cases for uber-go-lint-style

This skill provides a reproducible workflow and templates to create high-quality test cases for a single linter rule in the `uber-go-lint-style` repository. The goal is to produce tests that increase coverage, exercise edge cases, and ensure diagnostics match exactly what the analyzer reports.

## Purpose

- Create focused, maintainable test files for one analyzer.
- Encourage thorough coverage: positive/negative cases, edge inputs, multi-violation lines, and false-positive guards.
- Provide templates and example prompts to accelerate test creation.

## When to use

- After implementing or changing a rule in `rules/`.
- When you need to add regression tests for a reported bug.
- To expand coverage around corner cases and type-sensitive behavior.

## Inputs (what you need to start)

- Rule name (Analyzer `Name` as registered in `BuildAnalyzer()`), e.g. `builtin_name`.
- Exact diagnostic messages the analyzer emits (they must match `// want` comments exactly).
- The package under test (usually the testdata path under `testlintdata/<rulename>`).
- Optional: specific code samples or known failing inputs.

## Outputs

- One or more `.go` files in `testdata/src/testlintdata/<rulename>/` containing `// want "..."` comments marking expected diagnostics.
- Suggested additions to `plugin_test.go` (test function that runs the analyzer via `analysistest.Run`) if a new rule test function is needed.

## Step-by-step process (workflow)

1. Identify the analyzer to test.
   - Confirm its `Name` and the exact diagnostic `Message` strings in `rules/<rule>.go`.
2. Create the test directory: `testdata/src/testlintdata/<rulename>/` if not present.
3. Draft at least two files: `cases.go` (mixed good & bad examples) and `edge_cases.go` (targeted edge coverage).
4. For each violation you expect, add a `// want "..."` comment on the same line as the offending node.
5. Add GOOD examples with comments `// GOOD` and no `// want` so the test ensures no false positives.
6. Add multi-violation lines and overlapping cases to validate message matching.
7. Include type-sensitive cases (if analyzer uses TypesInfo) — e.g., interface wrappers, pointer vs value, aliased imports.
8. Add README or comment at top describing the intent of each test file.
9. Run `go test ./...` and iterate until `analysistest` matches exactly.

## Decision points / branching logic

- Does the analyzer use type info (`pass.TypesInfo`)?
  - Yes: include test files that import or declare types to exercise type resolution (aliases, wrapper types, renamed imports).
  - No: focus on structural AST patterns and comment/identifier placements.
- Is the diagnostic message parameterized or contextual (e.g., includes an identifier)?
  - Yes: ensure `// want` uses the exact runtime message produced by the analyzer for each case.
- Are false positives possible in macro-like patterns?
  - Add `//nolint:<rule>` examples where relevant and GOOD cases showing expected suppression.

## Quality criteria / completion checks

- Each expected violation has exactly one matching `// want` with the exact message string.
- There are explicit GOOD cases exercising similar code paths but not triggering the analyzer.
- Edge cases included: empty inputs, multiple violations on one line, shadowing, renamed imports, pointer vs value, interface implementations, unusual formatting.
- Tests exercise both AST-only and Type-aware behavior as appropriate.

## Templates

Test file header template (place at top of each test file):

```go
package <packagename>

// Test intent: <short description of what this file covers>

// BAD: <explain why this is bad>
// GOOD: <explain why this is fine>
```

A minimal BAD/GOOD example pattern:

```go
package samples

// BAD: uses builtin name 'error'
func bad(error string) { // want "identifier 'error' shadows a built-in, consider using a different name"
}

// GOOD: uses safe name
func good(msg string) {
}
```

Multi-violation one-line example (use multiple `// want` comments):

```go
func multi(error string, int int) { // want "identifier 'error' shadows a built-in, consider using a different name"
    // want "identifier 'int' shadows a built-in, consider using a different name"
}
```

Type-aware template (requires `LoadModeTypesInfo`):

```go
package samples

import "sync"

var mu sync.Mutex // GOOD: named mutex type is allowed

func bad() {
    var m sync.Mutex // want "use zero-value mutex or pointer receiver instead" // <-- example
}
```

## Example testcases checklist (use when drafting files)

- [ ] Simple BAD example (1 diagnostic)
- [ ] Simple GOOD example
- [ ] Multi-violation line
- [ ] Edge case: empty or minimal construct
- [ ] Type alias or renamed import
- [ ] Suppression via `//nolint:<rule>`
- [ ] Confusing but allowed case (to avoid false positives)
- [ ] Large input / repeated patterns (to bump coverage)

## How to add the test to `plugin_test.go`

If a test function doesn't already exist for your rule, add one modeled after existing tests:

```go
func Test<TitleCasedRule>(t *testing.T) {
    newPlugin, err := register.GetPlugin("uber-go-lint-style")
    require.NoError(t, err)

    plugin, err := newPlugin(nil)
    require.NoError(t, err)

    analyzers, err := plugin.BuildAnalyzers()
    require.NoError(t, err)

    // Find index of your analyzer in BuildAnalyzers() and use it here
    analysistest.Run(t, testdataDir(t), analyzers[<index>], "testlintdata/<rulename>")
}
```

Note: maintainers prefer tests that reuse the existing test runner structure — follow the ordering convention in `plugin.go`.

## Running and iterating locally

- Run the package tests:

```bash
go test ./...
```

- If `analysistest` fails, fix the `// want` strings or the test code until the messages match exactly.

## Example prompts to use with Copilot/Assistant

- "Create test cases for the `builtin_name` rule that cover shadowing, multi-violation lines, and aliasing."
- "Generate edge-case inputs for the `atomic` rule that include pointer and non-pointer types and renamed imports."
- "Add a test file with suppression examples for `mynewrule`, showing `//nolint` usage and GOOD cases."

## Tips and best practices

- Always copy the diagnostic message exactly from the analyzer implementation — small differences break `analysistest` matching.
- Put focused tests in separate files so it’s clear what each file validates.
- Prefer descriptive comments at the top of each test file describing intent.
- Keep test files small and focused; add more files rather than bundling unrelated checks together.
- When in doubt, add both a GOOD and BAD example that differ minimally — this helps catch false positives.

## Maintenance notes

- If the analyzer message changes, update all `// want` comments in the relevant testdata directory.
- When adding a new analyzer, ensure it is registered in `plugin.go` and placed at the expected index used by tests.

## Summary

This skill standardizes how to create, structure, and validate test cases for a single linter rule. Use the templates and checklist to ensure high coverage and robust edge-case handling.
