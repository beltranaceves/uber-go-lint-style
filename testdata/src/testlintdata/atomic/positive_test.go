package atomic_test

import (
	"sync/atomic"
)

type foo struct {
	running int32 // atomic
}

func (f *foo) start() {
	if atomic.SwapInt32(&f.running, 1) == 1 { // want "use go.uber.org/atomic instead of sync/atomic for operations on raw types"
		// already running
		return
	}
	// start the Foo
}

func (f *foo) isRunning() bool {
	return atomic.LoadInt32(&f.running) == 1 // want "use go.uber.org/atomic instead of sync/atomic for operations on raw types"
}
