// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
sval := T{Name: "foo"}

// inconsistent
sptr := new(T)
sptr.Name = "bar"
