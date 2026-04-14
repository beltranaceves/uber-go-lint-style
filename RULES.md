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

### `global_decl` — Top-level Variable Declarations

**What it detects:**
```go
var _s string = F()  // ❌ VIOLATION - redundant explicit type

func F() string { return "A" }
```

**Correct usage:**
```go
var _s = F() // ✅ OK - type inferred from initializer
```

Specify the type when the initializer's type differs from the declared type:

```go
type myError struct{}

func (myError) Error() string { return "error" }

func F() myError { return myError{} }

var _e error = F() // ✅ OK - explicit type required (widening to error)
```

**Why:** Repeating the type on a top-level `var` when the initializer already expresses the type is redundant and noisy. Prefer `var name = expr` for clarity; keep an explicit type only when you intentionally want a different (e.g., interface) type than the initializer provides.

**How the check works:**
The analyzer inspects top-level `var` declarations and, when a `ValueSpec` includes both an explicit type and an initializer, it uses type information (`pass.TypesInfo`) to compare the declared type with the initializer's type. If they are identical, the analyzer reports a diagnostic suggesting omitting the explicit type. Suppress with `//nolint:global_decl` for intentional exceptions.


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

### `goroutine_exit` — Wait for goroutines spawned by entrypoints

**What it detects:**

Bad:

```go
func main() {
		go func() {}() // ❌ VIOLATION - no visible wait
}

func TestMain(m *testing.M) {
		go func() {}() // ❌ VIOLATION - no visible wait
		os.Exit(m.Run())
}
```

Good:

```go
func main() {
		done := make(chan struct{})
		go func() {
				defer close(done)
		}()
		<-done // ✅ OK - main waits for goroutine
}

func init() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() { defer wg.Done() }()
		wg.Wait() // ✅ OK - init waits for goroutine
}
```

**Why:** Goroutines started directly from system-managed entrypoints (for
example `main()`, `init()`, or `TestMain`) have no caller that can join or wait
for them. Forgetting to synchronize such goroutines can cause the process or
tests to exit before background work completes.

**How the check works:**
- The rule is conservative by default and restricts analysis to entrypoints
	(`main`, `init`, `TestMain`) where missing waits are most likely to be
	problematic and easiest to reason about.
- It uses a layered approach for detection:
	- Lightweight AST/type heuristics: look for `go` statements in the entry
		function and nearby evidence of waiting in the same body (`wg.Wait()`,
		`close(...)` + receive, or direct `<-ch` expressions).
	- SSA-based interprocedural search: when SSA is available (via the
		`buildssa` pass) the analyzer finds `go` instructions and follows static
		callees (a depth-limited BFS) to detect whether control can reach a
		`sync.WaitGroup.Wait` method. This interprocedural traversal has a
		configurable depth limit (default 10) to avoid unbounded and expensive
		exploration.

**Scope & limitations:**
- Scope: by default checks only `main`, `init`, and `TestMain`. This keeps
	false positives low while catching the highest-risk omissions.
- SSA traversal follows only static callees discovered in SSA
	(`StaticCallee()`). It improves detection across packages for many common
	patterns but does not handle all indirect callsites (interface dispatch,
	reflection, or function-valued calls resolved via pointer analysis).
- The depth limit prevents worst-case performance on large callgraphs but may
	miss waits beyond that depth. You can increase the limit or enable pointer
	analysis for better precision at the cost of runtime.
- The rule still uses simple channel/`close`/receive heuristics to cover
	patterns where WaitGroup isn't used.

**Performance tradeoffs:**
- SSA and call-following improve accuracy but cost CPU/time. The analyzer
	avoids running heavy pointer-based callgraph analysis by default and uses a
	static-callee BFS with a depth limit. If you need full pointer-aware
	callgraphs, the analyzer can be extended to run `golang.org/x/tools/go/pointer`
	analysis in an opt-in mode (slower, but more complete).

**Suppressing:** Use `//nolint:goroutine_exit` to skip a specific site where a
goroutine without an obvious wait is intentional.


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


### `global_mut` — Avoid mutable package-level variables

**What it detects:**
```go
var counter = 0                // ❌ VIOLATION - mutable global
var _timeNow = time.Now        // ❌ VIOLATION - function-value global
var a, b = f(), g()            // ❌ VIOLATION - multi-name spec with runtime inits

const Version = "1.0"         // ✅ OK - consts are allowed
var ErrNotFound = errors.New("not found") // ✅ OK - sentinel error named Err*
var ExportedCounter = 0        // ✅ OK - exported package API allowed
```

**Why:**
Mutable package-level variables increase cognitive load, make code harder to
test, and can lead to surprising shared state and data races. Prefer passing
dependencies explicitly (dependency injection), scoping state behind types,
or exposing read-only package-level values via constructors or constants.

**How the check works:**
- The analyzer inspects top-level `var` declarations (AST `GenDecl` with `var`).
- It reports variables that are likely mutable: declared without an explicit
	exported name, or initialized with non-trivial expressions (function calls,
	composite literals, or basic literals).
- Exceptions:
	- Exported names (starting with an uppercase letter) are skipped because
		package APIs often require package-level values.
	- Sentinel errors named `Err...` whose initializer has the `error` type
		are allowed (detected using `pass.TypesInfo`).
- When a `ValueSpec` contains multiple names, the analyzer coalesces the
	result into a single diagnostic at the `var` site listing the flagged names.

**Implementation notes:**
- The analyzer uses `pass.TypesInfo` when available to detect `error`-typed
	initializers and should run under `LoadModeTypesInfo` for best accuracy.
- Diagnostics are reported via `analysis.Diagnostic` and are treated as
	violations by golangci-lint.

**Suppressing:** Use `//nolint:global_mut` to suppress the diagnostic for
specific cases where a package-level mutable variable is intentional.

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
- The analyzer inspects exported (`IsExported`) top-level functions and methods and only considers functions with three or more parameters.
- For accuracy it uses the `buildssa` pass to obtain SSA/CFG for the function and combines SSA information with lightweight AST checks:
	- It identifies the SSA basic blocks that reference each parameter (via SSA referrers).
	- It identifies return blocks in the function's CFG.
	- A parameter is considered optional if it is never referenced, or if there exists a path from the function entry to a return block that does not traverse any block that uses the parameter.
- The analyzer reports a diagnostic when the function has >= 3 parameters and at least one parameter is classified as optional by the SSA/CFG analysis.
- If SSA data is unavailable (rare in the plugin test harness), the analyzer falls back to a conservative syntactic report.
- This path-sensitive approach reduces false positives compared to simply counting parameters, but it is conservative about indirect uses (goroutines, reflection, stores to globals) and may be extended later for deeper interprocedural checks.

**Suppressing:** Use `//nolint:functional_option` to opt out for specific
cases where the pattern is undesirable.

