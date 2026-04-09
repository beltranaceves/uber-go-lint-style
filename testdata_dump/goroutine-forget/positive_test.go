// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
go func() {
  for {
    flush()
    time.Sleep(delay)
  }
}()
