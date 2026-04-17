package exit_once

import (
	"log"
	"os"
)

// BAD: multiple exit points in main
func main() {
	if len(os.Args) == 0 {
		log.Fatal("missing args")
	}
	if false {
		os.Exit(1)
	}
}

// GOOD: single exit point
func goodMain() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error { return nil }
