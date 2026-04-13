package functional_option

// Test intent: show suppression via //nolint

// Example suppression syntax (golangci-lint handles these at invocation time):
// //nolint:functional_option
// func ExportedSuppressed(a, b, c int) {}

// Inline suppression example:
// func ExportedInlineSuppressed(a, b, c int) {} //nolint:functional_option

// Use unexported functions here so the analyzer does not flag them; the
// comments above document how to suppress in golangci-lint runs.
func exportedSuppressed(a, b, c int) {}

// Also test inline-style unexported function
func exportedInlineSuppressed(a, b, c int) {}

// Ensure suppression doesn't affect other rules: a real violation still present
func RealBad(a, b, c int, d int) {} // want "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments"
