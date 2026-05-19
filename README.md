# uber-go-lint-style

[![Go Test](https://github.com/beltranaceves/uber-go-lint-style/actions/workflows/go-test.yml/badge.svg)](https://github.com/beltranaceves/uber-go-lint-style/actions/workflows/go-test.yml)
[![Coverage Status](https://codecov.io/gh/beltranaceves/uber-go-lint-style/branch/main/graph/badge.svg)](https://codecov.io/gh/beltranaceves/uber-go-lint-style)
[![Go Report Card](https://goreportcard.com/badge/github.com/beltranaceves/uber-go-lint-style)](https://goreportcard.com/report/github.com/beltranaceves/uber-go-lint-style)

A golangci-lint plugin for [Uber's Go Style Guide](https://github.com/uber-go/guide).

<p align="center">
  <img src="./assets/ACKCHYUALLY.png" alt="" width="300">
  <br>
  <!-- Logo by <a href="https://">origin</a> -->
</p>

> [!CAUTION]
> **Disclaimer**: this project contains significant amounts of auto-generated code, pending *thorough review*.

## Table of Contents

- [Overview](#overview)
- [Sample runs](#samples)
- [Installation](#installation)
	- [Prerequisites](#prerequisites)
	- [Setup Option 1: Automated Setup (Recommended)](#setup-option-1-automated-setup-recommended)
	- [Setup Option 2: Manual Configuration](#setup-option-2-manual-configuration)
- [Rules](#rules)
- [Development](#development)
	- [Project Structure](#project-structure)
	- [Adding a New Rule](#adding-a-new-rule)
	- [Running Tests](#running-tests)
- [Contributing](#contributing)
- [Resources](#resources)

## Overview

This is a custom linter that strives to enforce Uber's internal Go coding standards through static analysis. It's designed to catch style violations early and guide developers toward safer, more maintainable code patterns.

> [!WARNING]
> There are many subjective rules that are enforced through the use of heuristics. We recommend configuring it to report findings as warnings by default (see the "Installation" section for an example). Teams can opt into stricter severities when appropriate.

## Samples
<details>
<summary>Cadence - 19/05/2026</summary>

```bash
echo "Running Uber Go style linter (with golangci-lint)..."
Running Uber Go style linter (with golangci-lint)...
if [ ! -f "./custom-gcl" ]; then echo "Building custom golangci-lint with uber-go-lint-style plugin..."; golangci-lint custom || exit 1; fi; echo "Running Uber Go style golangci-lint..." ;./custom-gcl run --config .golangci.yml
Running Uber Go style golangci-lint...
common/clock/event_timer_gate.go:45:3: struct_embed: embedded field should be placed at the top of the struct (uber-go-lint-style)
                sync.RWMutex
                ^
common/clock/event_timer_gate.go:57:25: struct_field_key: use field names when initializing structs; specify fields like `Field: value` (uber-go-lint-style)
                fireTime:    time.Time{},
                                      ^
common/clock/event_timer_gate.go:79:2: var_scope: identifier 'active' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        active := t.currentTime.Before(t.fireTime)
        ^
common/clock/event_timer_gate.go:113:24: struct_field_key: use field names when initializing structs; specify fields like `Field: value` (uber-go-lint-style)
        t.fireTime = time.Time{}
                              ^
common/clock/event_timer_gate_test.go:44:7: struct_pointer: use &T instead of new T when initializing struct references (uber-go-lint-style)
        s := new(eventTimerGateSuite)
             ^
common/clock/ratelimiter.go:186:35: struct_field_key: use field names when initializing structs; specify fields like `Field: value` (uber-go-lint-style)
        _ Reservation = deniedReservation{}
                                         ^
common/clock/ratelimiter.go:228:2: var_scope: identifier 'newNow' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        newNow := r.timesource.Now() // caution: must be after acquiring the lock
        ^
common/clock/ratelimiter.go:280:2: var_scope: identifier 'res' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        res := r.limiter.ReserveN(now, 1)
        ^
common/clock/ratelimiter.go:358:5: var_scope: identifier 'err' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        if err := ctx.Err(); err != nil {
           ^
common/clock/ratelimiter.go:378:2: var_scope: identifier 'delay' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        delay := res.DelayFrom(now)
        ^
common/clock/ratelimiter.go:463:2: var_scope: identifier 'called' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        called := false
        ^
common/clock/ratelimiter_bench_test.go:97:2: decl_group: group adjacent var declarations into a single var block (uber-go-lint-style)
        var runSerial runType = func(b *testing.B, each func(int) bool) {
        ^
common/clock/ratelimiter_bench_test.go:100:7: var_scope: identifier 'i' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                for i := 0; i < b.N; i++ {
                    ^
common/clock/ratelimiter_bench_test.go:109:3: var_scope: identifier 'allowedPeriod' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                allowedPeriod := fmt.Sprintf(allowedPeriodFmt, "n/a")
                ^
common/clock/ratelimiter_bench_test.go:118:16: var_scope: identifier 'denied' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                var allowed, denied atomic.Int64
                             ^
common/clock/ratelimiter_bench_test.go:120:4: var_scope: identifier 'n' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                        n := 0
                        ^
common/clock/ratelimiter_bench_test.go:132:3: var_scope: identifier 'allowedPeriod' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                allowedPeriod := fmt.Sprintf(allowedPeriodFmt, "n/a")
                ^
common/clock/ratelimiter_bench_test.go:151:6: var_scope: identifier 'rl' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                                        rl := rate.NewLimiter(rate.Every(normalLimit), burst)
                                        ^
common/clock/ratelimiter_bench_test.go:157:6: var_scope: identifier 'rl' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                                        rl := NewRatelimiter(rate.Every(normalLimit), burst)
                                        ^
common/clock/ratelimiter_bench_test.go:163:6: var_scope: identifier 'ts' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                                        ts := NewMockedTimeSource()
                                        ^
common/clock/ratelimiter_bench_test.go:206:7: var_scope: identifier 'r' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                                                r := rl.Reserve()
                                                ^
common/clock/ratelimiter_bench_test.go:229:7: var_scope: identifier 'r' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                                                r := rl.ReserveN(now, 1)
                                                ^
common/clock/ratelimiter_bench_test.go:262:6: var_scope: identifier 'ts' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                                        ts := NewMockedTimeSource()
                                        ^
common/clock/ratelimiter_bench_test.go:318:8: var_scope: identifier 'rl' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                                                        rl := NewRatelimiter(limit, burst)
                                                        ^
common/clock/ratelimiter_comparison_test.go:86:7: var_scope: identifier 'testnum' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                for testnum := 0; !t.Failed() && time.Now().Before(deadline); testnum++ {
                    ^
common/clock/ratelimiter_comparison_test.go:120:5: var_scope: identifier 'seed' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                                seed := time.Now().UnixNano()
                                ^
common/clock/ratelimiter_comparison_test.go:256:3: var_scope: identifier 'round' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                round := make([]string, events)
                ^
common/clock/ratelimiter_comparison_test.go:269:3: var_scope: identifier 'set' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                set := rng.Intn(len(schedule) + 1)
                ^
common/clock/ratelimiter_comparison_test.go:320:2: var_scope: identifier 'compressed' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        compressed := NewRateLimiterWithTimeSource(compressedTS, limit, burst)
        ^
common/clock/ratelimiter_comparison_test.go:325:2: var_scope: identifier 'compressedReplay' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        compressedReplay := make([][]func(t *testing.T), rounds)
        ^
common/clock/ratelimiter_comparison_test.go:456:6: var_scope: identifier 'done' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                                        done := make(chan struct{})
                                        ^
common/clock/ratelimiter_comparison_test.go:547:2: var_scope: identifier 'maxLatency' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        maxLatency := maxDur(actual, wrapped, mocked)
        ^
common/clock/ratelimiter_comparison_test.go:548:2: var_scope: identifier 'minLatency' can be declared in the inner block to reduce its scope (uber-go-lint-style)
        minLatency := minDur(actual, wrapped, mocked)
        ^
common/clock/ratelimiter_comparison_test.go:573:3: var_scope: identifier 'assertNoWait' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                assertNoWait := func(what string, wait time.Duration) {
                ^
common/clock/ratelimiter_comparison_test.go:585:3: var_scope: identifier 'assertWaited' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                assertWaited := func(what string, wait time.Duration) {
                ^
common/clock/ratelimiter_test.go:42:3: var_scope: identifier 'name' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                name := name
                ^
common/clock/ratelimiter_test.go:45:4: var_scope: identifier 'ts' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                        ts := func() MockedTimeSource { return nil }
                        ^
common/clock/ratelimiter_test.go:293:5: struct_field_zero: omit zero-valued field "drainFirst" from struct literal; let Go set the zero value (uber-go-lint-style)
                                drainFirst: false,
                                ^
common/clock/ratelimiter_test.go:306:5: struct_field_zero: omit zero-valued field "allowed" from struct literal; let Go set the zero value (uber-go-lint-style)
                                allowed: 0,
                                ^
common/clock/ratelimiter_test.go:353:5: struct_field_zero: omit zero-valued field "drainFirst" from struct literal; let Go set the zero value (uber-go-lint-style)
                                drainFirst: false,
                                ^
common/clock/ratelimiter_test.go:399:5: struct_field_zero: omit zero-valued field "allowed" from struct literal; let Go set the zero value (uber-go-lint-style)
                                allowed:     0,
                                ^
common/clock/ratelimiter_test.go:447:11: type_assert: use the comma-ok form for type assertions (uber-go-lint-style)
                impl := rl.(*ratelimiter)
                        ^
common/clock/ratelimiter_test.go:451:12: type_assert: use the comma-ok form for type assertions (uber-go-lint-style)
                rimpl := r.(*allowedReservation)
                         ^
common/clock/sustain.go:25:1: decl_group: group import declarations into a single import block (uber-go-lint-style)
import "time"
^
common/clock/sustain.go:48:3: var_scope: identifier 'now' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                now := s.source.Now()
                ^
common/clock/sustain.go:68:3: var_scope: identifier 'now' can be declared in the inner block to reduce its scope (uber-go-lint-style)
                now := s.source.Now()
                ^
common/clock/sustain_test.go:111:4: struct_field_zero: omit zero-valued field "duration" from struct literal; let Go set the zero value (uber-go-lint-style)
                        duration: 0,
                        ^
common/clock/sustain_test.go:210:4: struct_field_zero: omit zero-valued field "duration" from struct literal; let Go set the zero value (uber-go-lint-style)
                        duration: 0,
                        ^
common/clock/timer_gate.go:47:3: struct_embed: add an empty line between embedded fields and regular fields (uber-go-lint-style)
                timeSource TimeSource
                ^
common/clock/timer_gate_test.go:43:7: struct_pointer: use &T instead of new T when initializing struct references (uber-go-lint-style)
        s := new(timerGateSuite)
             ^
50 issues:
* uber-go-lint-style: 50
make: *** [Makefile:6: uber_lint] Error 1
```
</details>

## Installation

### Prerequisites

- Go 1.23+
- golangci-lint 1.59.0+ ([Install docs](https://golangci-lint.run/usage/install/))


> [!TIP]
> If you are using a coding Agent (Claude Code, AmpCode, Cursor, Copilot, etc.), copy and paste this prompt:
> ```bash
> Fetch the install guide and follow it:
> curl -s https://raw.githubusercontent.com/beltranaceves/uber-go-lint-style/refs/heads/main/installation.md
> ```

Follow these steps:

### Setup Option 1: Automated Setup (Recommended)

Run the setup script to auto-generate configuration files:

```bash
go run github.com/beltranaceves/uber-go-lint-style/cmd/setup@latest
```

This creates:
- `.custom-gcl.yml` — Plugin configuration
- `.golangci.yml` — Linter settings
- `Makefile` — Build and run commands

Then simply:
```bash
make uber_lint
```

### Setup Option 2: Manual Configuration

If you prefer manual setup, follow these steps:

**Step 1: Create `.custom-gcl.yml`**

```yaml
version: v2.11.4

plugins:
	- module: 'github.com/beltranaceves/uber-go-lint-style'
		import: 'github.com/beltranaceves/uber-go-lint-style'
		version: 'latest'
```

**Step 2: Create a `.golangci.yml` to enable the plugin and rules**

```yaml
version: "2"

linters:
	default: none
	enable:
		- uber-go-lint-style
	settings:
		custom:
			uber-go-lint-style:
				type: "module"
				description: "Uber Go style guide linter"
				original-url: "github.com/beltranaceves/uber-go-lint-style"
				# Disabled rules provided as YAML text. By default exclude TodoRule.
				settings:
					disabled_rules_yaml: |
						- todo

severity:
	default: info
	rules:
		- linters:
				- uber-go-lint-style
			severity: warning
```

**Disabling plugin rules via YAML**

The `uber-go-lint-style` plugin accepts a YAML string in the `settings`
section inside your `.golangci.yml` to disable specific analyzers at runtime.
You can provide either a plain YAML list or a mapping with a `disabled:` (or
`disable:`) key. The entries must match the analyzer name returned by a rule's
`BuildAnalyzer()`.

Example (`.golangci.yml`):

```yaml
linters:
	settings:
		custom:
			uber-go-lint-style:
				settings:
					disabled_rules_yaml: |
						- TodoRule
						- AtomicRule
						- MapInitRule
```

**Step 3: Build the custom binary and run**

```bash
golangci-lint custom
./custom-gcl run --config .golangci.uber_style.yml
```

**Step 4: Add a Makefile (optional)**

To avoid running commands manually each time, add these targets to your `Makefile`:

```makefile
.PHONY: uber_lint
uber_lint: # Run Uber Go style linter (builds plugin if needed)
	$Q echo "Running Uber Go style linter (with golangci-lint)..."
	$Q if [ ! -f "./custom-gcl" ]; then echo "Building custom golangci-lint with uber-go-lint-style plugin..."; golangci-lint custom || exit 1; fi; echo "Running Uber Go style golangci-lint..." ;./custom-gcl run --config .golangci.uber_style.yml

.PHONY: uber_clean
uber_clean: # Clean Uber Go style linter artifacts
	$Q rm -f custom-gcl*
	$Q echo "Cleaned Uber Go style linter artifacts"

```

This automatically builds the binary on first run and caches it for subsequent runs. Then simply:

```bash
make uber_lint
```

## Rules

See [RULES.md](RULES.md) for full rule descriptions and examples.

## Development

### Project Structure

```
uber-go-lint-style/
├── plugin.go                # golangci-lint plugin entry point
├── plugin_test.go           # plugin tests
├── rules/                   # rule implementations (one file per rule)
├── testdata/                # testdata used by rule tests
├── cmd/                     # helper CLI tools (e.g., setup)
│   └── setup/               # setup command source
├── style_guide/             # generated and source docs for the style guide
│   └── rules/               # markdown source files for the guide
├── test-client/             # integration test client and examples
├── assets/                  # images and other assets
├── Makefile                 # convenience targets
├── installation.md          # installation instructions
└── RULES.md                 # rule descriptions and examples
```

### Adding a New Rule

> [!NOTE]
> If you are using coding Agents, or looking for more detailed guidance on rule structure, performance patterns, and testing conventions, there are two included [skills](.agents/skills/):
> - `.agents/skills/develop-linter-rules/SKILL.md` covers rule structure, analysis approaches, performance considerations, and examples.
> - `.agents/skills/create-linter-tests/SKILL.md` helps scaffold test cases and edge-case coverage to reduce boilerplate.

1. Create a new file in `rules/` (e.g., `rules/myrule.go`):

```go
package rules

import (
	"golang.org/x/tools/go/analysis"
)

type MyRule struct{}

func (r *MyRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "myrule",
		Doc: "enforce your style convention",
		Run: r.run,
	}
}

func (r *MyRule) run(pass *analysis.Pass) (any, error) {
	// Your linting logic here
	return nil, nil
}
```

2. Add test data in `testdata/src/testlintdata/myrule/`:

```go
package myrule_test

// Violations here
func bad() {
	undesirable code // want "error message"
}

// Good practices here  
func good() {
}
```

3. Add test in `plugin_test.go`:

```go
func TestMyRule(t *testing.T) {
	// Similar to existing test patterns
}
```

4. Register in `plugin.go`:

```go
func (f *PluginExample) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		(&rules.TodoRule{}).BuildAnalyzer(),
		(&rules.AtomicRule{}).BuildAnalyzer(),
		(&rules.MyRule{}).BuildAnalyzer(),  // Add here
	}, nil
}
```

### Running Tests

```bash
go test ./...
```

## Contributing

This project implements style rules from [Uber's Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md). When adding new rules:

1. Reference the specific style guideline being enforced
2. Document how the check works in the rule's `Doc` field
3. Provide comprehensive test cases (both good and bad patterns)
4. Keep rules focused and single-purpose

## Resources

- [uber-go/guide](https://github.com/uber-go/guide) — Uber's Go style guide
- [golangci-lint plugins](https://golangci-lint.run/docs/plugins/plugins-configuration/) — Custom plugin documentation
- Analysis tools:
  - [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)
  - [golang.org/x/tools/go/ssa](https://pkg.go.dev/golang.org/x/tools/go/ssa)
  - [go/ast](https://pkg.go.dev/go/ast)
  - [go/types](https://pkg.go.dev/go/types)

## License

This project is licensed under the Apache License, Version 2.0. See the
[LICENSE](LICENSE) file for details. A NOTICE file is included with
attribution where applicable.
