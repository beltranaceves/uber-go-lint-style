package goroutine_exit

import "sync"

func init() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // want "goroutine started in main/init/TestMain must have a way to wait for it to exit"
		defer wg.Done()
	}()
}
