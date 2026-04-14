package goroutine_forget

import "time"

// BAD: fire-and-forget goroutine with infinite loop
func bad() {
	go func() { // want "fire-and-forget goroutine contains an infinite loop; provide stop signaling and a way to wait for the goroutine to exit"
		for {
			flush()
			time.Sleep(time.Second)
		}
	}()
}

// BAD: for true infinite loop
func badTrue() {
	go func() { // want "fire-and-forget goroutine contains an infinite loop; provide stop signaling and a way to wait for the goroutine to exit"
		for true {
			flush()
		}
	}()
}

// BAD: named function started as goroutine with an infinite loop
func badNamed() {
	go worker() // want "fire-and-forget goroutine contains an infinite loop; provide stop signaling and a way to wait for the goroutine to exit"
}

func worker() {
	for {
		flush()
	}
}

// GOOD: goroutine can be signaled to stop and waited on
func good() {
	var (
		stop = make(chan struct{})
		done = make(chan struct{})
	)
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				flush()
			case <-stop:
				return
			}
		}
	}()

	// elsewhere
	close(stop)
	<-done
}

func flush() {}
