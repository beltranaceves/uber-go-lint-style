// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

import uberatomic "go.uber.org/atomic"

// Example 1
type foo struct {
	running uberatomic.Bool
}

func (f *foo) start() {
	if f.running.Swap(true) {
		// already running
		return
	}
	// start the Foo
}

func (f *foo) isRunning() bool {
	return f.running.Load()
}
