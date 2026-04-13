package goroutine_exit

func init() {
	ch := make(chan struct{})
	go func() {
		ch <- struct{}{}
	}()
	<-ch
}
