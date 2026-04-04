// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
type foo struct {
  running int32  // atomic
}

func (f* foo) start() {
  if atomic.SwapInt32(&f.running, 1) == 1 {
     // already running…
     return
  }
  // start the Foo
}

func (f *foo) isRunning() bool {
  return f.running == 1  // race!
}
