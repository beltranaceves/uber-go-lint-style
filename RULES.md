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


