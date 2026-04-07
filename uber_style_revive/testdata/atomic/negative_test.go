// Auto-generated test case for atomic rule
// Negative = should PASS lint (Good code)

package testdata

import uberatomic "go.uber.org/atomic"

// Example: Using go.uber.org/atomic (GOOD)
type Counter struct {
	value uberatomic.Int64
}

func (c *Counter) Increment() {
	c.value.Add(1)
}

func (c *Counter) Get() int64 {
	return c.value.Load()
}
