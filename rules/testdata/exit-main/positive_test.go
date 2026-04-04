// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
func main() {
  body := readFile(path)
  fmt.Println(body)
}

func readFile(path string) string {
  f, err := os.Open(path)
  if err != nil {
    log.Fatal(err)
  }

  b, err := io.ReadAll(f)
  if err != nil {
    log.Fatal(err)
  }

  return string(b)
}
