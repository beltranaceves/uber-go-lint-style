// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
// foo.go

const (
  defaultPort = 8080
  defaultUser = "user"
)

// bar.go

func Bar() {
  defaultPort := 9090
  ...
  fmt.Println("Default port", defaultPort)

  // We will not see a compile error if the first line of
  // Bar() is deleted.
}
