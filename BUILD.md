# Building and Testing Custom Revive Rules

This document explains how to build the Uber Go Style Guide rules into the revive linter and test them.

## Architecture Overview

```
uber-go-lint-style/
├── rules/                    # Custom rule implementations
│   ├── *.go                  # Individual rule files (58 rules)
│   └── testdata/             # Test cases for each rule
│       └── [rule-name]/      # positive_test.go, negative_test.go
├── revive.toml               # Linter configuration
└── internal/                 # Build tools
    ├── extract/              # Test case extraction
    ├── generate/             # Rule skeleton generation
    └── agents/               # Agent framework
```

## Prerequisites

1. **Clone revive repository**:
```bash
git clone git@github.com:mgechev/revive.git
cd revive
```

2. **Copy rules**:
```bash
# Copy your rules directory to revive
cp -r /path/to/uber-go-lint-style/rules/* rule/
```

## Building Revive with Custom Rules

### Option 1: Local Development

1. Add your rules to the revive rule package:
```bash
# In revive directory
cp -r ../uber-go-lint-style/rules/*.go rule/
```

2. Register rules in rule/default_config.go or add to revive.toml:
```bash
# Your revive.toml already has the rules configured
```

3. Build revive:
```bash
go build -o revive .
```

### Option 2: Using the Test Framework

Run the test framework to verify test case structure:
```bash
go test ./rules/... -v -run TestAllRules
```

This verifies:
- Each rule has positive and negative test cases
- Test files are properly structured

### Option 3: Running Lint Tests

Once revive is built with your rules:
```bash
# Test a single rule
./revive -config ../uber-go-lint-style/revive.toml ./rules/testdata/defer-clean/positive_test.go

# Expected: Should report lint failures (positive = Bad code)

# Test negative cases
./revive -config ../uber-go-lint-style/revive.toml ./rules/testdata/defer-clean/negative_test.go

# Expected: Should report no failures (negative = Good code)
```

## Agent Framework for Parallel Rule Implementation

The project includes an agent framework for parallel rule implementation:

### Initialize Agents
```bash
go run internal/cmd/init_agents.go
```

This creates status files in `.agent-status/[rule-name].json`

### Agent Loop Process

Each agent follows this process:

1. **Read test cases**: Load positive/negative examples from `rules/testdata/[rule]/`
2. **Implement rule**: Add logic to `rules/[rule-name].go`
3. **Run tests**: Execute `go test ./rules/...` to verify
4. **Iterate**: If tests fail, modify implementation and retry

### Monitoring Progress

Check agent status:
```bash
# View individual rule status
cat .agent-status/defer-clean.json

# List all pending rules
ls -la .agent-status/
```

## Rule Implementation Template

Each rule should implement the `lint.Rule` interface:

```go
package rules

import (
    "github.com/mgechev/revive/lint"
)

type RuleNameRule struct{}

// Name returns the rule name (matches revive.toml entry)
func (r *RuleNameRule) Name() string {
    return "rule-name"
}

// Apply runs the rule against the provided file
func (r *RuleNameRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
    var failures []lint.Failure
    
    // Walk file.AST to find violations
    // For each violation:
    //   failures = append(failures, lint.Failure{
    //       Failure:     "description",
    //       Node:        astNode,
    //       Confidence: 1.0,
    //   })
    
    return failures
}
```

## Testing Individual Rules

### Manual Testing

```bash
# Build revive first
cd revive && go build -o revive .

# Test defer-clean rule
./revive -config ../uber-go-lint-style/revive.toml \
    -formatter friendly \
    ../uber-go-lint-style/rules/testdata/defer-clean/positive_test.go
```

Expected output for positive_test.go (Bad code):
```
rules/testdata/defer-clean/positive_test.go:8:3: use defer for unlock (defer-clean)
```

Expected output for negative_test.go (Good code):
```
(no output - code follows style guide)
```

### Automated Testing

To fully automate the agent loop, you would:

1. Create an agent that reads the rule's test cases
2. Implements the rule logic
3. Runs revive on each test case
4. Validates positive cases FAIL and negative cases PASS
5. Iterates until all tests pass

## Current Status

- **58 rules** have test cases extracted from the style guide
- **58 empty rule skeletons** created in `rules/*.go`
- **58 agent status files** tracking implementation progress
- **Test framework** validates test case structure

## Next Steps

To complete the implementation loop:

1. Build revive with custom rules
2. Implement each rule's logic (walking the AST to detect violations)
3. Use the agent framework to track progress
4. Iterate until all test cases pass

The hardest part is step 2 - implementing the AST traversal logic for each rule. Each rule requires understanding what patterns violate the style guide and how to detect them in Go code.