# Architecture Overview

## System Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    User Interaction Layer                        │
├──────────────────────┬──────────────────────┬──────────────────┤
│  Standalone CLI      │  golangci-lint       │   Test Harness   │
│  (uber-go-lint)      │   Integration        │   (TestAllRules) │
└──────────────┬───────┴──────────────┬───────┴──────────────┬────┘
               │                      │                      │
               └──────────────────────┼──────────────────────┘
                                      │
┌──────────────────────────────────────▼──────────────────────────┐
│              Linter Runner (linter/runner.go)                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ • Creates revive.Linter instance                        │   │
│  │ • Loads rule configuration                             │   │
│  │ • Executes linting pipeline                            │   │
│  │ • Formats output (friendly, simple, json)              │   │
│  └─────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────▲──────────────────────────┘
                                       │
┌──────────────────────────────────────▼──────────────────────────┐
│            Rule Registry (rules/init.go)                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ NewRule(name) → lint.Rule                              │   │
│  │                                                         │   │
│  │ "atomic"       → AtomicRule{}                          │   │
│  │ "error-wrap"   → ErrorWrapRule{}                       │   │
│  │ "error-name"   → ErrorNameRule{}                       │   │
│  │ "struct-embed" → StructEmbedRule{}                     │   │
│  │ "global-mut"   → GlobalMutRule{}                       │   │
│  └─────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────▲──────────────────────────┘
                                       │
┌──────────────────────────────────────▼──────────────────────────┐
│              Individual Rule Implementations                     │
│  ┌────────────────┬────────────────┬────────────────────────┐   │
│  │ atomic.go      │ error_wrap.go  │ error_name.go          │   │
│  │ ┌────────────┐ │ ┌────────────┐ │ ┌──────────────────┐   │   │
│  │ │Name()      │ │ │Name()      │ │ │Name()            │   │   │
│  │ │Apply()     │ │ │Apply()     │ │ │Apply()           │   │   │
│  │ └────────────┘ │ └────────────┘ │ └──────────────────┘   │   │
│  │                │                │                        │   │
│  │ struct_embed   │ global_mut     │                        │   │
│  │ .go            │ .go            │                        │   │
│  └────────────────┴────────────────┴────────────────────────┘   │
│                                                                 │
│  All implement: lint.Rule interface                            │
│  ├─ Name() string                                              │
│  └─ Apply(file *lint.File, args lint.Arguments) []lint.Failure│
└──────────────────────────────────────▲──────────────────────────┘
                                       │
┌──────────────────────────────────────▼──────────────────────────┐
│                   Test Data (testdata/)                         │
│  ┌────────────────┬────────────────┬────────────────────────┐   │
│  │  atomic/       │  error-wrap/   │  error-name/           │   │
│  │├ positive_t.go │├ positive_t.go │├ positive_test.go      │   │
│  │└ negative_t.go │└ negative_t.go │└ negative_test.go      │   │
│  │                │                │                        │   │
│  │  struct-embed/ │  global-mut/   │                        │   │
│  │├ positive_t.go │├ positive_test │                        │   │
│  │└ negative_t.go │└ negative_test │                        │   │
│  └────────────────┴────────────────┴────────────────────────┘   │
│                                                                 │
│  Convention: <ruleName>/{positive,negative}_test.go            │
│  Positive:   Bad code → should have ≥1 lint failures          │
│  Negative:   Good code → should have 0 lint failures          │
└──────────────────────────────────────────────────────────────────┘
```

## Data Flow - Testing Pipeline

```
┌─────────────────┐
│   go test .     │
│   ./rules -v    │
└────────┬────────┘
         │
         ▼
┌──────────────────────────────────────┐
│ TestAllRules()                       │
│ 1. Scan testdata/ directory          │
│ 2. Discover rule subdirectories      │
└────────┬─────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────┐
│ For each rule (parallel t.Run)       │
│ - testRule("atomic")                 │
│ - testRule("error-wrap")             │
│ - testRule("error-name")             │
└────────┬─────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────┐
│ For each rule: testRule(t, name)     │
│ 1. Verify rule exists in init.go     │
│ 2. Load positive_test.go fixtures    │
│ 3. Load negative_test.go fixtures    │
└────────┬─────────────────────────────┘
         │
    ┌────┴────┐
    │          │
    ▼          ▼
┌────────┐  ┌────────┐
│Positive│  │Negative│
│Tests   │  │Tests   │
└────┬───┘  └───┬────┘
     │          │
     ▼          ▼
┌─────────────────────────────────────┐
│ runRuleOnFixture(rule, fixture)     │
│ 1. Create revive.Linter             │
│ 2. Configure rule                   │
│ 3. Lint fixture file                │
│ 4. Collect failures                 │
└────────┬────────────────────────────┘
         │
    ┌────┴─────┐
    │           │
    ▼           ▼
┌─────────┐  ┌─────────┐
│Assert   │  │Assert   │
│len>0 ✓  │  │len==0 ✓ │
└─────────┘  └─────────┘
     │           │
     └───┬───────┘
         │
         ▼
    ┌─────────┐
    │ PASS ✓  │
    └─────────┘
```

## Data Flow - CLI Usage

```
┌─────────────────────────────────┐
│ ./uber-go-lint ./myproject      │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ main.go (cmd/uber-go-lint)      │
│ 1. Parse flags (-format, paths) │
│ 2. Create linter.Runner         │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ Runner.LintPaths(paths)         │
│ 1. Get all rules from init.go   │
│ 2. Create revive.Linter         │
│ 3. Execute Linter.Lint()        │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ revive.Lint(patterns, rules)    │
│ Returns: failures channel       │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ For each failure (from channel) │
│ - Collect failure details       │
│ - File, line, column, message   │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ Format failures                 │
│ - friendly (default)            │
│ - simple                        │
│ - json                          │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ Print output + exit code        │
│ Exit(0) if no failures          │
│ Exit(1) if failures found       │
└─────────────────────────────────┘
```

## Integration with golangci-lint

```
┌─────────────────────────────────────┐
│ golangci-lint run -c .golangci.yml  │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│ Load .golangci-uber.yml             │
│ ┌─────────────────────────────────┐ │
│ │ linters-settings:               │ │
│ │   custom:                       │ │
│ │     uber-go-style:              │ │
│ │       path: ./bin/uber-go-lint  │ │
│ │       original-url: ...         │ │
│ └─────────────────────────────────┘ │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│ golangci-lint spawns subprocess:    │
│ ./bin/uber-go-lint ./...            │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│ CLI tool runs (same as standalone)  │
│ Outputs failures in known format    │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│ golangci-lint parses output         │
│ Aggregates with other linters       │
│ Reports results to user             │
└─────────────────────────────────────┘
```

## Rule Implementation Example: atomic.go

```go
package rules

type AtomicRule struct{} // Stateless rule

// Name identifies this rule
func (r *AtomicRule) Name() string {
    return "atomic"
}

// Apply analyzes one file's AST
func (r *AtomicRule) Apply(
    file *lint.File,
    args lint.Arguments,
) []lint.Failure {
    
    var failures []lint.Failure
    
    // Traverse AST nodes
    for _, imp := range file.AST.Imports {
        // Check if importing sync/atomic
        if path == "sync/atomic" {
            failures = append(failures, lint.Failure{
                Failure:    "Use go.uber.org/atomic",
                Node:       imp,
                Confidence: 1.0,
            })
        }
    }
    
    return failures
}
```

## Extensibility Points

```
┌─────────────────────────────────────────────────────┐
│ Extend with New Rules                               │
│ 1. Create rules/new_rule.go (implements Rule)      │
│ 2. Create testdata/new-rule/{positive,negative}    │
│ 3. Register in rules/init.go (3 functions)         │
│ 4. Run tests (auto-discovered)                     │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│ Extend CLI Output Formats                           │
│ Add new formatter in internal/linter/runner.go     │
│ • formatXML()                                       │
│ • formatSarif()                                     │
│ • formatMarkdown()                                 │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│ Extend Configuration System                         │
│ Load TOML/YAML from file                           │
│ • Select which rules to enable                     │
│ • Set rule-specific arguments                      │
│ • Control output formatting                        │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│ Extend with Pre/Post Hooks                         │
│ • Before linting (setup)                           │
│ • After linting (reporting)                        │
│ • Error handling                                   │
└─────────────────────────────────────────────────────┘
```

## Key Design Principles

1. **Convention over Configuration**
   - Rule discovery by directory name
   - Test fixtures follow naming convention
   - No manual registration needed for tests

2. **Minimal Dependencies**
   - Only depends on `revive`
   - Self-contained implementation
   - Easy to understand and modify

3. **Composable Architecture**
   - Standalone CLI tool
   - Works as golangci-lint plugin
   - Library usable in other tools

4. **Interface-Based Design**
   - Implements `lint.Rule` interface
   - Rules are stateless
   - Easy to test in isolation

5. **Parallel Testing**
   - Multiple fixtures run concurrently
   - Fast feedback loop
   - Good for CI/CD integration
