package goroutine_exit

// This goroutine is started in a non-entry function and should not be flagged
// by the analyzer which only checks main/init/TestMain.
func Start() {
	go func() {}()
}
