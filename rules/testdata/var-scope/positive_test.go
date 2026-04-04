// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
err := os.WriteFile(name, data, 0644)
if err != nil {
 return err
}

// Example 2
if data, err := os.ReadFile(name); err == nil {
  err = cfg.Decode(data)
  if err != nil {
    return err
  }

  fmt.Println(cfg)
  return nil
} else {
  return err
}

// Example 3
const (
  _defaultPort = 8080
  _defaultUser = "user"
)

func Bar() {
  fmt.Println("Default port", _defaultPort)
}
