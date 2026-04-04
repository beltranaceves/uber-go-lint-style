# Uber Go Lint Style - Auto-generation Project

## Overview

This project aims to automatically generate revive lint rules from Uber's Go Style Guide. The workflow involves:
1. Extracting test cases from style guide markdown files
2. Creating empty rule implementations for each style rule
3. Building a test suite to validate rules
4. Running parallel agents to implement rules until they pass all test cases

## Current State

- **61 style rule markdown files** in `style_guide/rules/`
- **1 example rule template** in `rules/example_rule.go`
- **Standard revive.toml** with 54 enabled rules and 5 custom rule placeholders

---

## PHASE 1: Autogenerate Test Cases

### 1.1 Parse Style Guide Markdown Files

**Goal**: Extract all code examples from markdown files and convert them to testable Go files.

**Steps**:
- [ ] Create a markdown parser that reads each `style_guide/rules/*.md` file
- [ ] Extract code blocks tagged as `go` from "Bad" and "Good" columns in tables
- [ ] Create directory structure: `rules/testdata/[rule-name]/`
- [ ] Save "Bad" examples as `positive_test.go` (rule should detect violations)
- [ ] Save "Good" examples as `negative_test.go` (rule should pass)
- [ ] Handle edge cases:
  - Multi-line code blocks
  - Code blocks with comments explaining context
  - Code blocks spanning multiple table rows

**Output Structure**:
```
rules/testdata/
├── error-once/
│   ├── positive_test.go  # Should trigger lint failure
│   └── negative_test.go  # Should pass lint
├── error-wrap/
│   ├── positive_test.go
│   └── negative_test.go
├── defer-clean/
│   ├── positive_test.go
│   └── negative_test.go
└── ... (one directory per style rule)
```

### 1.2 Validate Test Case Structure

**Goal**: Ensure extracted test cases are valid Go code.

**Steps**:
- [ ] Write a script to check each extracted file compiles
- [ ] Fix any syntax errors in extracted code
- [ ] Add necessary imports for compilation
- [ ] Ensure code is self-contained (no external dependencies)

---

## PHASE 2: Create Empty Revive Rules

### 2.1 Generate Rule Skeleton Files

**Goal**: Create empty rule implementations for each style rule.

**Steps**:
- [ ] Create `rules/[rule-name].go` for each of the 61 style rules
- [ ] Each file follows the template:
```go
package rules

import (
    "github.com/mgechev/revive/lint"
)

type RuleNameRule struct{}

// Name returns the rule name
func (r *RuleNameRule) Name() string {
    return "rule-name"
}

// Apply runs the rule against the provided file
func (r *RuleNameRule) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
    var failures []lint.Failure
    // TODO: Implement rule logic
    return failures
}
```

### 2.2 Register Rules in revive.toml

**Goal**: Add all generated rules to the configuration.

**Steps**:
- [ ] Generate TOML entries for each rule
- [ ] Add to `revive.toml` under custom rules section:
```toml
[[rule]]
name = "rule-name"
severity = "warning"
description = "Description from style guide"
```

---

## PHASE 3: Build Test Suite

### 3.1 Create Test Runner Framework

**Goal**: Build infrastructure to run each rule against its test cases.

**Steps**:
- [ ] Create `rules/rules_test.go` with test helpers
- [ ] Create `testutil/testdata.go` for loading test cases
- [ ] Implement test runner that:
  - Loads positive/negative test files
  - Runs revive with the specific rule
  - Validates positive files FAIL (rule detects issue)
  - Validates negative files PASS (rule doesn't flag)

**Test Structure**:
```go
// rules/rule_name_test.go
func TestRuleName(t *testing.T) {
    // Test positive cases - should fail lint
    lintutils.RunLinterTest(t, "rule-name", "positive", true)
    // Test negative cases - should pass lint
    lintutils.RunLinterTest(t, "rule-name", "negative", false)
}
```

### 3.2 Create Test Utilities

**Goal**: Provide reusable utilities for rule testing.

**Steps**:
- [ ] Create `rules/testutil/load.go` - loads testdata files
- [ ] Create `rules/testutil/lint.go` - runs revive on test code
- [ ] Create `rules/testutil/assert.go` - assertions for test results

### 3.3 Validate Framework

**Goal**: Ensure test framework works with existing example.

**Steps**:
- [ ] Run tests on `example_rule.go` template
- [ ] Verify tests pass/fail as expected
- [ ] Fix any framework issues

---

## PHASE 4: Parallel Agent Execution

### 4.1 Define One-Agent-Per-Rule Infrastructure

**Goal**: Set up system where each rule gets its own agent to implement.

**Architecture**:
```
┌─────────────────────────────────────────────────────────┐
│                    Main Orchestrator                     │
├─────────────────────────────────────────────────────────┤
│  rules/error-once/agent.go  │  rules/error-wrap/agent.go │
│  ┌─────────────────────┐    │  ┌─────────────────────┐   │
│  │ Implements rule    │    │  │ Implements rule    │   │
│  │ Runs tests         │    │  │ Runs tests         │   │
│  │ Iterates until OK  │    │  │ Iterates until OK  │   │
│  └─────────────────────┘    │  └─────────────────────┘   │
├─────────────────────────────────────────────────────────┤
│              Shared: testdata/, testutil/               │
└─────────────────────────────────────────────────────────┘
```

**Steps**:
- [ ] Create agent framework in `agents/agent_framework.go`
- [ ] Each agent:
  1. Reads its rule's test cases
  2. Implements the rule logic
  3. Runs tests
  4. If fails: modifies implementation, runs tests again
  5. If passes: marks complete, reports status

### 4.2 Define Agent Communication Protocol

**Goal**: Agents communicate status to orchestrator.

**Steps**:
- [ ] Define status enum: `PENDING`, `IN_PROGRESS`, `PASSED`, `FAILED`
- [ ] Create status file: `.agent-status/[rule-name].json`:
```json
{
  "rule": "error-once",
  "status": "IN_PROGRESS",
  "attempts": 3,
  "last_error": "expected failure but passed"
}
```
- [ ] Implement orchestrator that:
  - Spawns agents in parallel (or batch of N)
  - Monitors status files
  - Retries failed agents
  - Reports completion

### 4.3 Implement Agent Loop Logic

**Goal**: Each agent implements an iterative improvement loop.

**Agent Algorithm**:
```go
func (a *Agent) Run() {
    for attempt := 0; attempt < MAX_ATTEMPTS; attempt++ {
        // 1. Implement rule logic
        a.ImplementRule()

        // 2. Run tests
        result := a.RunTests()

        // 3. Check results
        if result.AllPassed() {
            a.ReportSuccess()
            return
        }

        // 4. Analyze failures, modify implementation
        a.AnalyzeAndFix(result.Failures())
    }
    a.ReportFailed()
}
```

### 4.4 Create Agent Task Prompts

**Goal**: Provide each agent with clear instructions.

**Steps**:
- [ ] Create prompt template for each agent:
```
You are implementing revive rule: [rule-name]
Reference: [link to style guide section]

Your task:
1. Read test cases in rules/testdata/[rule-name]/
2. Implement the rule in rules/[rule-name].go
3. Run tests to validate

Test Cases:
- Positive (should FAIL lint): [list files]
- Negative (should PASS lint): [list files]

Requirements:
- Rule must detect all violations in positive test cases
- Rule must NOT flag any issues in negative test cases
- Follow revive rule interface pattern
```

### 4.5 Run Parallel Agent Sessions

**Goal**: Execute all agents in parallel.

**Steps**:
- [ ] Launch 61 parallel agent tasks (one per rule)
- [ ] Monitor progress via status files
- [ ] Handle failures with retry logic
- [ ] Aggregate final results

---

## Implementation Priority Order

### Sprint 1: Infrastructure Setup
1. ✅ Parse markdown → extract test cases
2. ✅ Generate rule skeletons
3. ✅ Build test runner framework

### Sprint 2: Validation
4. ✅ Validate with 5 placeholder rules in revive.toml
5. ✅ Fix framework issues

### Sprint 3: Agent Execution
6. ✅ Run agents for all 61 rules
7. ✅ Collect results
8. ✅ Report completion status

---

## File Manifest

### New Directories
```
rules/
├── testdata/
│   ├── error-once/{positive,negative}_test.go
│   ├── error-wrap/{positive,negative}_test.go
│   ├── defer-clean/{positive,negative}_test.go
│   └── ... (61 directories)
├── testutil/
│   ├── load.go
│   ├── lint.go
│   └── assert.go
└── * _test.go (61 test files)

agents/
├── agent_framework.go
├── status.go
└── orchestrator.go
```

### New Files
```
TODO.md (this file)
.make/ (makefile targets for automation)
.agent-status/ (status tracking)
```

---

## Success Criteria

1. **Test Cases**: All 61 style rules have positive and negative test cases
2. **Rules**: All 61 rules have empty skeleton implementations
3. **Test Suite**: All rules can be tested via `go test ./rules/...`
4. **Agent Execution**: All rules implemented correctly (pass their test cases)
5. **Integration**: Rules usable via `revive -config revive.toml`

---

## Notes

- **Revive Interface**: Rules implement `lint.Rule` with `Name()` and `Apply()` methods
- **Test Format**: "Bad" code → should trigger failure; "Good" code → should pass
- **Agent Parallelism**: Run 10-20 agents concurrently to avoid overwhelming system
- **Timeout**: Each agent has 10-minute timeout per attempt, max 5 attempts