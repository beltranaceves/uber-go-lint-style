// Auto-generated test case for struct-embed rule
// Negative = should PASS lint (Good code)

package testdata

// Example: Explicit field names with basic types (GOOD)
type GoodConfig struct {
	Data    string          // Explicit field name
	Timeout int             // Explicit field name
	Retries uint32
}

// Another example with custom embedding
type CustomType struct {
	Value int
}

type GoodUser struct {
	Custom CustomType      // Explicit named embedding
	Name   string
}
