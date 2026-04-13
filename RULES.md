## Rules

### `todo` — Require author in TODO comments

**What it detects:**
```go
// TODO: fix this  // ❌ VIOLATION - no author
// TODO(): fix this // ❌ VIOLATION - malformed

// TODO(alice): fix this  // ✅ OK - has author
```

**Why:** Unattributed TODOs can be lost or unmaintained. Requiring an author ensures accountability and provides context for future developers.

### `atomic` — Use go.uber.org/atomic for raw types

**What it detects:**
```go
var counter int32
atomic.StoreInt32(&counter, 1)  // ❌ VIOLATION - raw type

val := atomic.LoadInt32(&counter)  // ❌ VIOLATION - returns raw type
```

**Correct usage:**
```go
counter := atomic.NewInt32(0)
counter.Store(1)  // ✅ OK - type-safe wrapper
val := counter.Load()
```

**Why:** The `sync/atomic` package operates on raw types, making it easy to forget atomic operations. `go.uber.org/atomic` provides type-safe wrappers that prevent accidental non-atomic access.

**How the check works:**
The rule inspects the function signature of `sync/atomic` calls and flags those that take or return raw types (int32, int64, uint32, uint64, uintptr). These should be replaced with equivalent operations from `go.uber.org/atomic`.

### `builtin_name` — Avoid shadowing predeclared identifiers

**What it detects:**
```go
func example(error string) { } // ❌ VIOLATION - 'error' shadows builtin

type T struct {
	string string // ❌ VIOLATION - 'string' shadows builtin
}
```

**Why:** Using predeclared identifiers as variable, parameter, receiver, or field names makes code confusing and can shadow standard names, reducing readability and making searches harder.

**How the check works:**
The analyzer dynamically retrieves predeclared identifiers from `go/types.Universe` and inspects `GenDecl`, function parameters/receivers, and struct fields via the AST to report shadowing occurrences.

### `error_name` — Error naming conventions

**What it detects:**
```go
var BrokenLink = errors.New("broken")        // ❌ VIOLATION - exported error should be prefixed Err
var notFound = fmt.Errorf("not found")      // ❌ VIOLATION - unexported error should be prefixed err

type NotFound struct{}                         // ❌ VIOLATION - implements Error() but lacks 'Error' suffix
func (n NotFound) Error() string { return "" }

// Correct examples:
var ErrCouldNotOpen = errors.New("could not open")
var errInternal = errors.New("internal")
type ResolveError struct{}
func (r ResolveError) Error() string { return "" }
```

**Why:** Exported package-level error variables should be discoverable and consistent (`Err` prefix) so callers can match them with `errors.Is`. Unexported package errors should follow a parallel `err` prefix to signal package-local use. Custom error types should end with the `Error` suffix to make their intent obvious and simplify error type matching with `errors.As`.

**How the check works:**
The analyzer inspects package-level `var` declarations and uses `pass.TypesInfo` to detect variables of the built-in `error` type. It enforces `Err`/`err` prefixes for exported and unexported variables respectively. It also looks at named types and checks whether they implement an `Error() string` method; if so, it requires the type name to end with `Error`. The rule runs under the plugin's `LoadModeTypesInfo` so type information is available.

### `channel_size` — Prefer unbuffered or size one channels

**What it detects:**
```go
c := make(chan int, 64) // ❌ VIOLATION - unusual buffer size
```

**Why:** Non-trivial channel buffer sizes should be deliberate and documented; most channels are unbuffered or sized to one for simple handoff semantics.

**How the check works:**
It finds `make(chan T, N)` calls and flags capacities that are not `0` or `1`. Non-literal capacities are reported conservatively for review.

### `container_capacity` — Preallocate container capacity when populating in loops

**What it detects:**
```go
m := make(map[string]V)
for _, v := range src {
	m[k] = v // ❌ VIOLATION - preallocate map capacity
}

s := make([]T, 0)
for _, v := range src {
	s = append(s, v) // ❌ VIOLATION - preallocate slice capacity
}
```

**Why:** Preallocating capacity for maps and slices when the size is known avoids repeated allocations and improves performance.

**How the check works:**
The analyzer records `make` calls that omit capacity and then scans loop bodies (range/for) for map index assignments or `append` calls that populate those containers; it reports at the original `make` site.

### `container_copy` — Copy slices/maps at API boundaries

**What it detects:**
```go
func (s *Store) Set(items []T) {
	s.items = items // ❌ VIOLATION - stores caller's slice directly
}

func (s *Store) Items() []T {
	return s.items // ❌ VIOLATION - returns internal slice without copying
}
```

**Why:** Returning or storing caller-owned slices or maps without copying can leak internal state, cause accidental sharing, and increase risk of data races.

**How the check works:**
Inspects function bodies to detect assignments that store parameter slices/maps into receiver fields and returns that expose receiver-owned slices/maps; reports where a copy should be made.

### `decl_group` — Encourage grouping similar declarations

**What it detects:**
```go
import "a"
import "b" // ❌ VIOLATION - import declarations should be grouped

const A = 1
const B = 2 // ✅ grouped suggestion when related

var x int = 1
var y int = 2 // ✅ grouped suggestion when same explicit type
```

**Why:** Grouping related `import`, `const`, `var`, and `type` declarations improves readability and follows Go conventions.

**How the check works:**
It's a conservative AST-only analyzer that looks for adjacent single-spec `GenDecl`s. It always suggests grouping multiple single `import` declarations. For top-level `const`, `var`, and `type` it recommends grouping only when declarations clearly share an explicit type, literal kind, or `iota` usage to avoid false positives. Function-local adjacent `var` declarations are recommended to be grouped even if unrelated, per style guidance.

### `defer_clean` — Use `defer` to clean up resources such as files and locks

**What it detects:**
```go
p.Lock()
if p.count < 10 {
	p.Unlock() // ❌ VIOLATION - non-deferred unlock before an early return
	return p.count
}

f, _ := os.Open("file")
f.Close() // ❌ VIOLATION - non-deferred close
```

**Correct usage:**
```go
p.Lock()
defer p.Unlock()

f, _ := os.Open("file")
defer f.Close()
```

**Why:** Missing cleanup calls (for example, `Unlock` or `Close`) are easy to miss across multiple return paths. `defer` keeps the cleanup adjacent to the acquisition and reduces the chance of leaks or forgotten unlocks while having negligible runtime overhead in typical functions.

**How the check works:**
The analyzer looks for selector calls named `Unlock`, `RUnlock`, or `Close` and reports those that are not directly used in a `defer` statement. It is conservative and may produce false positives in intentional manual-cleanup patterns (for example, unlocking inside tight loops); such cases can be suppressed with `//nolint:defer_clean`.

### `else_unnecessary` — Avoid unnecessary `else` when both branches set the same variable

**What it detects:**
```go
var a int
if b {
	a = 100
} else {
	a = 10
}
```

**Why:** Initializing the variable to the `else` value and keeping a single `if` branch is clearer and shorter:
```go
a := 10
if b {
	a = 100
}
```

**How the check works:**
The analyzer inspects `if` statements with a plain `else` block and reports cases where both the `if` and `else` bodies consist of a single assignment to the same identifier. It reports a diagnostic at the `if` site with a suggestion to initialize the variable before the `if` and remove the `else` block. Complex assignments, declarations (":="), or multi-statement branches are ignored to avoid false positives.

### `embed_public` — Avoid Embedding Types in Public Structs

**What it detects:**
```go
// BAD: exported struct embedding exported type
type ConcreteList struct {
	*AbstractList // ❌ VIOLATION - leaks implementation detail
}

// GOOD: use a private field and explicit delegate methods
type ConcreteList struct {
	list *AbstractList
}
```

**Why:** Embedding a public type in a public struct exposes implementation details, constrains future changes (removing or replacing the embedded type is a breaking change), and makes documentation harder to read.

**How the check works:**
This AST-based analyzer looks for exported (`type` names starting with an uppercase letter) struct declarations that contain anonymous (embedded) fields whose type name is exported. It reports the embedded field position with a clear diagnostic. Cases can be suppressed with `//nolint:embed_public` when embedding is intentional.

### `enum_start` — Start enums at one

**What it detects:**
```go
type Operation int

const (
	Add Operation = iota // ❌ VIOLATION - starts at 0
	Subtract
)
```

**Why:** Starting enumerations at 1 prevents the zero value from being a valid enum member by accident. The zero value is the default for uninitialized variables; reserving zero as an invalid or sentinel value avoids subtle bugs where the zero value accidentally matches a meaningful enum variant.

**How the check works:**
- It inspects top-level `const` groups that use `iota` and are associated with a named integer type (e.g., `type T int`).
- If the first enumerator in the group evaluates to `0` (either implicitly via no initializer or explicitly via `iota`, `0`, or `iota + 0`), the analyzer reports a diagnostic recommending starting the enum at `1` (for example, by using `iota + 1` or adding an explicit `Unknown`/`Unset` sentinel at zero).

**Detection heuristic:** This rule includes heuristics to reduce false positives: it only applies to const groups tied to named integer types and ignores unrelated `iota` uses. It also recognizes common cases where zero is intentional (and can be suppressed with `//nolint:enum_start` or an explanatory comment). Because some complex constant expressions are hard to evaluate statically, the analyzer relies on conservative checks rather than full constant evaluation; review cases that the analyzer flags to confirm intent.


### `error_once` — Handle errors once

**What it detects:**
```go
u, err := getUser(id)
if err != nil {
    log.Printf("could not get user %q: %v", id, err)
    return err // ❌ VIOLATION - logged and returned
}

// Acceptable: wrap with %w and return
if err := doThing(); err != nil {
    return fmt.Errorf("doThing: %w", err) // ✅ OK - wrapped return
}

// Acceptable: log and degrade (no return)
if err := emitMetrics(); err != nil {
    log.Printf("metrics failed: %v", err) // ✅ OK - log and continue
}
```

**Why:**
Logging an error and returning it in the same place causes duplicated handling
and noisy logs when callers also log or handle the error. Prefer returning the
error (optionally wrapped with `%w`) so higher-level callers control logging
and handling, or log and recover locally without returning the error.

**How the check works:**
- AST-only analyzer that finds `if err != nil { ... }` blocks where a call with
  a common logging method name (Printf, Errorf, Infof, etc.) appears and the
  error identifier is returned in the same block.
- Treats `fmt.Errorf("...%w...", err)` as a safe wrapped return and does not flag it.
- Conservative by design: it matches common logging method names and local
  logger method calls, but may miss custom logging helpers or logging that
  happens outside the `if` body. Suppress with `//nolint:error_once` when
  logging-and-return is intentional.

### `exit_main` — Call `os.Exit`, `log.Fatal`, and `panic` only in `main()`

**What it detects:**
```go
func helper() {
	if err != nil {
		log.Fatal(err) // ❌ VIOLATION - only allowed inside main()
	}
}

func anotherHelper() {
	os.Exit(1) // ❌ VIOLATION - only allowed inside main()
}

func panicker() {
	panic("boom") // ❌ VIOLATION - panic used as program exit
}

func main() {
	if err != nil {
		log.Fatal(err) // ✅ OK
	}
}
```

**Why:** Exiting from non-`main` functions makes control flow non-obvious, is
hard to test (it may terminate tests), and skips deferred cleanup. `panic`
should not be used as a program-exit mechanism — prefer returning an error so
callers (and `main`) can decide how to handle failures.

**How the check works:**
The analyzer walks function bodies and uses `pass.TypesInfo` to resolve selector
expressions. It flags calls to `os.Exit`, `log` package functions with names
starting with `Fatal`, and plain `panic(...)` invocations when they occur
outside of the `main` function in package `main`. Files ending in `_test.go`
are ignored. Suppress with `//nolint:exit_main` when termination from a helper
is intentional.

### `function_order` — Group and order functions for readability

**What it detects:**
```go
// Bad: type declared after methods
func (s *something) Stop() {}

type something struct{}

// Bad: constructor appears after methods
func newSomething() *something { return &something{} }

// Bad: methods of A interleaved with B
func (a *A) One() {}
func (b *B) First() {}
func (a *A) Two() {}

// Bad: exported methods appear after unexported
func (c *C) unexported() {}
func (c *C) Exported() {}

// Bad: caller declared after callee
func (d *D) Callee() {}
func (d *D) Caller() { d.Callee() }
```

**Why:**
Grouping functions by receiver and ordering them (constructors, exported methods,
then helpers) improves local reasoning, makes call flow easier to follow, and
reduces search/grep friction when maintaining a type.

**How the check works:**
- Ensures top-level `type`, `const`, and `var` declarations appear before any
	function declarations in a file.
- Detects a `NewX`/`newX` constructor and requires it to appear immediately
	after the corresponding `type X` declaration and before `X`'s methods.
- Groups methods by the receiver and enforces that methods for a receiver form
	a contiguous block (no interleaving with other receivers or package-level
	functions).
- Enforces exported methods appear before unexported ones within each receiver
	block.
- Conservative call-order check: for methods on the same receiver, if method
	A contains a direct selector call `r.B()` on the receiver identifier and A
	appears after B, the analyzer reports a diagnostic asking to declare A
	before B. This is a syntactic, conservative check — it only considers
	statically visible selector calls on the receiver variable and does not
	attempt to resolve interface dispatch or function-value calls.

**Detection details & caveats:**
- The rule operates per-file; it does not reorder or compare across multiple
	files in a package.
- The call-order logic uses syntactic receiver-identifier matching (e.g., `s.`)
	and may miss or conservatively ignore calls performed via interfaces, alias
	variables, or reflection. For stricter resolution, type-aware or SSA-based
	analysis would be required.
- The analyzer reports the first offending out-of-order top-level declaration
	for the types-before-functions rule to avoid noisy output.

**Suppressing:** Use `//nolint:function_order` to skip checks in exceptional
cases where ordering must differ for clarity or initialization reasons.

### `functional_option` — Prefer functional options for expandable APIs

**What it detects:**
```go
func Open(addr string, cache bool, logger *zap.Logger) (*Connection, error) // ❌ VIOLATION - 3+ params

func OpenWithOpts(addr string, opts ...Option) (*Connection, error) // ✅ OK - functional options
```

**Why:**
The functional options pattern improves API ergonomics and future-proofing for
exported constructors and public functions that already take several
parameters. It makes optional arguments explicit, avoids breaking changes when
adding new options, and can make defaults and configuration clearer to callers.

**How the check works:**
- The analyzer is AST-only and runs per-file during the analysis `Run` pass.
- It inspects exported (`IsExported`) top-level functions and methods.
- It counts parameters by summing the parameter names (an anonymous field
	counts as one). If the total is 3 or more the analyzer reports a diagnostic
	at the function identifier recommending the functional options pattern.
- This rule is conservative and purely syntactic: it does not use type
	information and does not try to distinguish which parameters are optional.

**Suppressing:** Use `//nolint:functional_option` to opt out for specific
cases where the pattern is undesirable.

