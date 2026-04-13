package goroutine_exit

func main() {
	go func() {}() // want "goroutine started in main/init/TestMain must have a way to wait for it to exit"
}
