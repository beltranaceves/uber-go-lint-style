// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
s, err := store.New()
if err != nil {
    return fmt.Errorf(
        "new store: %w", err)
}
