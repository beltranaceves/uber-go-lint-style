package goroutine_init

// BAD: init spawns a goroutine
func init() {
	go doWork() // want "do not start goroutines in init"
}

func doWork() {}

// BAD: init calls a function that starts a goroutine indirectly
func init() {
	startWorker() // want "do not start goroutines in init"
}

func startWorker() {
	go doWork()
}

// GOOD: no goroutine in init
func init_good() {
	// nothing
}
