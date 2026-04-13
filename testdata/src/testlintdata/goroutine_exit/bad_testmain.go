package goroutine_exit

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	go func() {}() // want "goroutine started in main/init/TestMain must have a way to wait for it to exit"
	os.Exit(m.Run())
}
