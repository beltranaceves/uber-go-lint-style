package main

import (
	"log"
	"os"
)

func helper() {
	_, err := os.Open("/does/not/exist")
	if err != nil {
		log.Fatal(err) // want "call to os.Exit or log.Fatal"
	}
}

func anotherHelper() {
	os.Exit(1) // want "call to os.Exit or log.Fatal"
}

func panicker() {
	panic("boom") // want "panic should not be used to exit programs"
}

func main() {
	_, err := os.Open("/does/not/exist")
	if err != nil {
		log.Fatal(err)
	}
}
