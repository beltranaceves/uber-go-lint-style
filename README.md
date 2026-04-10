# uber-go-lint-style

Set of custom rules for the Go revive linter following [Uber's Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md#linting).

## Project Structure

```
uber-go-lint-style/
├── rules/                    # Custom revive rule implementations
│   ├── example_rule.go      # Template for implementing rules
│   ├── README.md            # Guidelines for developing custom rules
│   └── (custom rules)       # Additional rule implementations
├── docs/                    # Documentation
│   ├── DEVELOPMENT.md       # Custom rule development guide
│   └── USAGE.md            # Revive usage and integration guide
├── style_guide/            # Uber Go style guide rules documentation
├── revive.toml             # Revive configuration file with standard and custom rules
└── README.md               # This file
```

## Quick Start

### Installation

1. Install revive:
   ```bash
   go install github.com/mgechev/revive@latest
   ```

2. Clone/navigate to this repository

### Running Linter

```bash
# Lint entire project
revive ./...

# Lint with this config
revive -config revive.toml ./...
```

### Developing Custom Rules

1. Read [`rules/README.md`](rules/README.md) for implementation guidelines
2. Use `rules/example_rule.go` as a template
3. Reference [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) for detailed development guide
4. Add rule configuration to `revive.toml`

## golangci-lint Plugin Interface

Custom linters compatible with golangci-lint must implement the `register.LinterPlugin` interface:

- **`BuildAnalyzers()`** — Returns `[]*analysis.Analyzer` defining the linter's analyzers and their execution function
- **`GetLoadMode()`** — Declares load mode: `LoadModeSyntax` (AST-only) or `LoadModeTypesInfo` (full type information)

Each analyzer reports issues via `pass.Report(analysis.Diagnostic)` with:
- **`Pos`** — Exact token position (line/column) where the issue occurs
- **`Message`** — Description of the issue
- **`Category`** — Linter name for grouping

The linter must be registered via `register.Plugin("name", New)` and follow the standard `golang.org/x/tools/go/analysis` framework. See [example.go](example.go) for a complete implementation.

## Configuration

The `revive.toml` file contains:
- All standard revive rules
- Placeholders for custom Uber-style rules
- Global linter settings

See [`docs/USAGE.md`](docs/USAGE.md) for configuration details.

## Documentation

- **[rules/README.md](rules/README.md)**: Rule development structure and guidelines
- **[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)**: In-depth custom rule development guide
- **[docs/USAGE.md](docs/USAGE.md)**: Revive usage, integration, and troubleshooting

## Resources

- [Revive Linter](https://github.com/mgechev/revive)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
