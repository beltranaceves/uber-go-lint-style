// Auto-generated test case for atomic rule
// Positive = should FAIL lint (Bad code)

package testdata

import "sync/atomic"

// Example: Using sync/atomic directly (BAD)
type Counter struct {
	value int64
}

func (c *Counter) Increment() {
	atomic.AddInt64(&c.value, 1)
}

func (c *Counter) Get() int64 {
	return atomic.LoadInt64(&c.value)
}
