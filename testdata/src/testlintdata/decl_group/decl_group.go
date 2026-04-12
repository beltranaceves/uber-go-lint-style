package decl_group

// Imports: should be grouped
import "fmt" // want "group import declarations into a single import block"

// Top-level consts: related (same literal kind) -> want grouping
const A = 1 // want "group related const declarations into a single const block"
const B = 2

// Top-level consts: unrelated (different literal kinds) -> no diagnostic
const C = "x"
const D = 3

// Top-level vars: related (same type) -> want grouping
var Va int = 1 // want "group related var declarations into a single var block"
var Vb int = 2

// Top-level vars: unrelated -> no diagnostic
var Vc = 1
var Vd = "s"

// Type declarations: related -> want grouping
type T1 int // want "group related type declarations into a single type block"
type T2 int

// Function-local vars: adjacent var declarations -> want grouping
func fn() {
	var x = 1 // want "group adjacent var declarations into a single var block"
	var y = 2
	// unrelated statements should break runs
	_ = fmt.Sprintf("%d %d", x, y)
	var z = "s" // want "group adjacent var declarations into a single var block"
	var w = "t"
	_ = fmt.Sprintf("%s%s", z, w)
}
