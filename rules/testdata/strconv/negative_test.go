// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
for i := 0; i < b.N; i++ {
  s := strconv.Itoa(rand.Int())
}
