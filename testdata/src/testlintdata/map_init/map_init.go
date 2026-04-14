package map_init

// BAD: empty map composite literal — prefer make
var (
	m1 = map[int]string{} // want "use make for empty maps"
)

func badShort() {
	m := map[string]int{} // want "use make for empty maps"
	_ = m
}

// GOOD: use make for empty maps
func goodMake() {
	m := make(map[string]int)
	_ = m
}

// GOOD: map literal with fixed elements is preferred
func goodLiteral() {
	m := map[string]int{
		"a": 1,
		"b": 2,
	}
	_ = m
}
