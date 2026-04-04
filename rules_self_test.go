package main

import (
	"testing"

	_ "github.com/beltranaceces/uber-go-lint-style/rules"
)

// This file allows running `go test` to execute the rules test suite
// Tests are in the rules/ subdirectory

func TestMain(m *testing.M) {
	m.Run()
}
