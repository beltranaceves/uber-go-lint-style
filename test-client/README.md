# Test Client for uber-go-lint-style

This is a test project that simulates a real-world scenario where a client uses the `uber-go-lint-style` golangci-lint plugin.

## Purpose

- **Validate plugin integration** with golangci-lint
- **Test rule detection** in a realistic setup
- **Verify configuration** and plugin loading
- **Simulate client usage** scenario

## Quick Start

### Run the linter:

```bash
cd test-client
make
```

That's it! 

- **First run**: Takes ~1-2 minutes to build the custom golangci-lint binary with the plugin
- **Subsequent runs**: Instant (binary is cached)

### View results

You should see violations reported by the `uber-go-lint-style` plugin rules:

```
main.go:8:2: TODO comment has no author (uber-go-lint-style)
main.go:14:2: use go.uber.org/atomic instead of sync/atomic for operations on raw types (uber-go-lint-style)
examples_test.go:14:2: use go.uber.org/atomic instead of sync/atomic for operations on raw types (uber-go-lint-style)
...
```

### Other commands

```bash
make clean    # Remove cached plugin binary (forces rebuild next time)
make help     # Show available targets
```

## How It Works

The Makefile automates the build process:

1. **First run** (`make`): Detects that `custom-gcl` doesn't exist
2. Runs `golangci-lint custom` to build the plugin (reads `.custom-gcl.yml` for config)
3. Runs `./custom-gcl run` to lint your code
4. **Subsequent runs**: Binary is cached, so linting is instant

To force a rebuild (after plugin changes):
```bash
make clean && make
```

## Behind the Scenes

**`.custom-gcl.yml`** — Tells golangci-lint where to find your plugin:
```yaml
plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    path: ../  # Local path for development
```

**`.golangci.yml`** — Enables the plugin and its rules:
```yaml
linters:
  enable:
    - uber-go-lint-style
```

**`custom-gcl` binary** — The result: golangci-lint compiled with your plugin built in

## Testing Against GitHub

To test against a released version instead of local:

1. Push and create a GitHub release (`v0.1.0`)
2. Update `.custom-gcl.yml`:
   - Change `path: ../` to `version: v0.1.0`
3. Run `make clean && make`

**That's it.** The rest is automatic.

## Troubleshooting

**"command not found: golangci-lint"**
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Plugin or rules not detected**
- Check `.custom-gcl.yml` points to correct path (`../`)
- Run `make clean` and rebuild

**Want to use published version instead of local**

Update `.custom-gcl.yml`:
```yaml
plugins:
  - module: 'github.com/beltranaceves/uber-go-lint-style'
    version: v0.1.0  # Use released version instead of path
```

Then rebuild: `make clean && make`
