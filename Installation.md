# Installation Guide for LLM Agents

If you're an LLM agent helping set up **uber-go-lint-style**, welcome! Follow these steps to install and configure the linter plugin.

## Prerequisites

Verify these dependencies are installed:

```bash
go version          # Go 1.23+
golangci-lint --version  # 1.59.0+
```

If golangci-lint is not installed, see the [install guide](https://golangci-lint.run/usage/install/).

## Setup Steps

### Option 1: Automated Setup (Recommended)

The fastest path. Runs a setup script that generates all config files:

```bash
go run github.com/beltranaceves/uber-go-lint-style/cmd/setup@latest
```

Or if testing locally:

```bash
go run ./cmd/setup
```

This creates three files:
- `.custom-gcl.yml` — Plugin configuration
- `.golangci.yml` — Linter settings  
- `Makefile` — Build and run commands

Then run the linter:

```bash
make uber_lint
```

### Option 2: Manual Configuration

If the user prefers full control:

**Step 1: Create `.custom-gcl.yml`**

```yaml
version: v1.59.0

plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.1
```

For local development, replace `version: v0.1.1` with `path: /path/to/uber-go-lint-style`.

**Step 2: Create `.golangci.yml`**

```yaml
version: "1"

linters:
  disable-all: true
  enable:
    - uber-go-lint-style

linters-settings:
  custom:
    uber-go-lint-style:
      type: "module"
      description: "Uber Go style guide linter"
      original-url: "github.com/beltranaceves/uber-go-lint-style"

severity:
  default-severity: error
  rules:
    - linters:
        - uber-go-lint-style
      severity: warning
```

**Step 3: Build and run**

```bash
golangci-lint custom
./custom-gcl run ./...
```

## Verification

Confirm the setup works:

```bash
# Check that the custom binary was created
ls -la custom-gcl

# Run a test
./custom-gcl run ./...
# or if using Makefile:
make uber_lint
```

Expected output: Linting results with uber-go-lint-style rules applied.

## Troubleshooting

**Binary not found after `golangci-lint custom`**

```bash
# Rebuild explicitly
golangci-lint custom --rebuild
```

**Plugin not loading**

Verify `.custom-gcl.yml` syntax:

```bash
cat .custom-gcl.yml
```

Ensure the plugin version/path is correct and the module exists.

**Rules not firing**

Check `.golangci.yml` has `uber-go-lint-style` enabled under `linters.enable`.

## Next Steps

- See [RULES.md](RULES.md) for individual rule documentation
- Review [style_guide/style.md](style_guide/style.md) for the full Uber Go Style Guide
- Integrate into CI/CD by adding `make uber_lint` to your pipeline
