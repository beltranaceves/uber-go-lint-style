// Auto-generated test case for struct-embed rule
// Positive = should FAIL lint (Bad code)

package testdata

// Example: Embedding basic types (BAD)
type BadConfig struct {
	string                  // Embedding basic type
	int                     // Embedding basic type
	Timeout int             // Named field is good
}

// Another example
type BadUser struct {
	error                   // Embedding error type
	Name string
}
