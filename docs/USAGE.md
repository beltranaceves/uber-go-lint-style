# Using Revive with Custom Rules

## Setup

1. **Install revive**:
   ```bash
   go install github.com/mgechev/revive@latest
   ```

2. **Configuration file**: The `revive.toml` in the repository root is automatically detected by revive

## Running Revive

### Basic Usage

```bash
# Lint entire project
revive ./...

# Lint specific directory
revive ./mypackage

# Lint specific file
revive ./mypackage/file.go
```

### With Configuration

```bash
# Use specific config file
revive -config revive.toml ./...

# Show all rules (including disabled)
revive -list-rules
```

### Output Options

```bash
# JSON output
revive -formatter json ./...

# SARIF format (for GitHub/GitLab integration)
revive -formatter sarif ./...

# TAP format
revive -formatter tap ./...

# JUnit format
revive -formatter junit-xml ./...
```

## Configuration Reference

### Rule Severity Levels

- `error`: Serious violations that should be fixed
- `warning`: Non-critical issues to be aware of
- `note`: Informational issues

### Common Rule Arguments

Many rules accept configuration arguments:

```toml
[[rule]]
name = "cognitive-complexity"
severity = "warning"
arguments = [7]  # Maximum complexity threshold

[[rule]]
name = "cyclomatic"
severity = "warning"
arguments = [15]  # Maximum cyclomatic complexity

[[rule]]
name = "var-naming"
severity = "warning"
arguments = ["ID"]  # Acceptable abbreviations
```

## Integration Examples

### GitHub Actions

Create `.github/workflows/lint.yml`:

```yaml
name: Lint
on: [push, pull_request]

jobs:
  revive:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - run: go install github.com/mgechev/revive@latest
      - run: revive ./...
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
revive -config revive.toml ./...
if [ $? -ne 0 ]; then
    echo "Revive linting failed"
    exit 1
fi
```

### VS Code Integration

1. Install [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go)
2. Configure in `.vscode/settings.json`:
   ```json
   {
     "go.lintTool": "revive",
     "go.lintOnSave": "package"
   }
   ```

### GoLand/IntelliJ IDEA

1. Go to Settings → Go → Linter
2. Select "Revive"
3. Set configuration file path to `revive.toml`
4. Enable "Run linter on code inspection"

## Customizing the Configuration

### Enable/Disable Rules

To enable a rule:
```toml
[[rule]]
name = "rule-name"
severity = "warning"
```

To disable a rule, remove or comment out its entry.

### Set Global Severity

```toml
# At the top of revive.toml
severity = "warning"  # Default severity for all rules
```

### Exclude Files/Patterns

```toml
ignoreFailures = false  # Set to true to not fail on violations
```

### Per-Rule Configuration

```toml
[[rule]]
name = "unhandled-error"
severity = "warning"
arguments = [
    "fmt.Fprintf",
    "fmt.Fprint",
    "fmt.Printf"
]  # List of functions to check
```

## Troubleshooting

### Rule not working?

1. Check the rule is enabled in `revive.toml`
2. Verify spelling: `revive -list-rules`
3. Check severity level isn't too lenient
4. Ensure arguments are correctly formatted

### Unwanted violations?

1. Add rule to `revive.toml` with appropriate arguments
2. Use ignore comments (if supported):
   ```go
   // nolint:rule-name
   func myFunction() {
   }
   ```

### Performance issues?

1. Reduce number of enabled rules
2. Disable expensive rules (cognitive-complexity, cyclomatic)
3. Check for circular dependencies
4. Profile with: `revive -with-id -v ./...`

## Best Practices

1. **Start permissive**: Enable fewer rules initially, add gradually
2. **Team consensus**: Review configuration with team before enforcing
3. **Document**: Add comments explaining non-obvious rule configurations
4. **CI/CD**: Always run revive in continuous integration
5. **Fix gradually**: Phase in strict rules to avoid overwhelming team
6. **Keep updated**: Periodically update revive for new rules

## See Also

- [Revive Documentation](https://github.com/mgechev/revive)
- [Uber Go Style Guide](https://github.com/uber-go/guide)
- [Custom Rules Development](./DEVELOPMENT.md)
