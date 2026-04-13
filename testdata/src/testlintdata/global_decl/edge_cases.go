package global_decl

// Edge cases: ensure analyzer is top-level only and handles composite literals

// GOOD: function-local var with explicit type should NOT be flagged
func local() {
	var localInt int = 1
	_ = localInt
}

// BAD: composite literal with same explicit type
var m map[string]int = map[string]int{"a": 1} // want "omit the explicit type in top-level var; use var name = expr instead"

// GOOD: composite literal without explicit type
var mGood = map[string]int{"b": 2}

// GOOD: explicit type with no initializer (should be allowed)
var z int
