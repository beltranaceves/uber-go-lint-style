package atomic_test

// This code doesn't import sync/atomic, so it passes the linter.
// Ideally it would use go.uber.org/atomic instead, but for linting purposes
// we just need to show code that doesn't use sync/atomic.

type bar struct {
	running bool
}

func (b *bar) start() {
	if b.running {
		// already running
		return
	}
	b.running = true
	// start the bar
}

func (b *bar) isRunning() bool {
	return b.running
}
