// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
func isActive(now, start, stop int) bool {
  return start <= now && now < stop
}

// Example 2
func poll(delay int) {
  for {
    // ...
    time.Sleep(time.Duration(delay) * time.Millisecond)
  }
}

poll(10) // was it seconds or milliseconds?

// Example 3
// {"interval": 2}
type Config struct {
  Interval int `json:"interval"`
}
