// Auto-generated test case for global-mut rule
// Positive = should FAIL lint (Bad code)

package testdata

// Example: Mutable global variables (BAD)
var (
	Config map[string]string = make(map[string]string) // Exported mutable global
	State  int                                           // Exported mutable global
)

// Another mutable global
var Logger interface{} // Could be reassigned

func UpdateConfig(key, value string) {
	Config[key] = value
}

func SetState(s int) {
	State = s
}
