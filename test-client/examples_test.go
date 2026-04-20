package examples

import (
	"sync/atomic"
)

// ATOMIC RULE EXAMPLES
// =====================

// BAD: Using sync/atomic with raw int32
func badAtomicInt32() {
	var counter int32
	atomic.StoreInt32(&counter, 1)    // use go.uber.org/atomic instead of sync/atomic for operations on raw types
	val := atomic.LoadInt32(&counter) // use go.uber.org/atomic instead of sync/atomic for operations on raw types
	_ = val
}

// BAD: Using sync/atomic with raw uint64
func badAtomicUint64() {
	var id uint64
	atomic.StoreUint64(&id, 100)      // use go.uber.org/atomic instead of sync/atomic for operations on raw types
	current := atomic.LoadUint64(&id) // use go.uber.org/atomic instead of sync/atomic for operations on raw types
	_ = current
}

// BAD: CompareAndSwap with raw type
func badAtomicCompareSwap() {
	var flag int32
	atomic.CompareAndSwapInt32(&flag, 0, 1) // ❌ should use go.uber.org/atomic
}

// TODO: add better error handling
func todoWithoutAuthor() {
	// This triggers the TODO rule
}

// TODO(alice): add validation logic
func todoWithAuthor() {
	// This is OK - has author
}

// GOOD EXAMPLES (commented out - uncomment to test with go.uber.org/atomic)
// ========================================================================

// import atomicpkg "go.uber.org/atomic"
//
// func goodAtomicBool() {
// 	running := atomicpkg.NewBool(false)
// 	running.Store(true)         // ✅ type-safe
// 	isRunning := running.Load() // ✅ type-safe
// 	_ = isRunning
// }
//
// func goodAtomicInt32() {
// 	counter := atomicpkg.NewInt32(0)
// 	counter.Store(1)             // ✅ type-safe
// 	val := counter.Load()         // ✅ type-safe
// 	_ = val
// }
