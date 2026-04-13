package goroutine_exit

func init() {
	done := make(chan struct{})
	go func() {
		defer close(done)
	}()
	<-done
}
