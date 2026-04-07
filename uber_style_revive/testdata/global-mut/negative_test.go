// Auto-generated test case for global-mut rule
// Negative = should PASS lint (Good code)

package testdata

import "sync"

// Example: Using constants or immutable structures (GOOD)
const (
	DefaultTimeout = 30
	MaxRetries     = 3
)

// Unexported mutable globals (internal to package) are acceptable
var (
	mutex    sync.Mutex // Unexported mutex
	internal map[string]string
)

func GetConfig(key string) (string, bool) {
	return internal[key], false
}

func UpdateConfig(key, value string) {
	mutex.Lock()
	defer mutex.Unlock()
	internal[key] = value
}

// Using sync.Mutex for thread-safe mutations
type ConfigManager struct {
	mu     sync.Mutex
	config map[string]string
}
