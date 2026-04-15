package mutex_zero_value

import "sync"

// BAD: pointer via new
func badNew() {
	mu := new(sync.Mutex) // want "avoid pointer to sync.Mutex; use zero-value sync.Mutex instead"
	_ = mu
}

// BAD: pointer var
var globalMu *sync.Mutex // want "avoid pointer to sync.Mutex; use zero-value sync.Mutex instead"

// BAD: address composite literal
func badAddr() {
	m := &sync.Mutex{} // want "avoid pointer to sync.Mutex; use zero-value sync.Mutex instead"
	_ = m
}

// BAD: embedded mutex
type SMap struct {
	sync.Mutex // want "do not embed sync.Mutex; use a named field instead"
	data       map[string]string
}

// GOOD: named field
type GoodSMap struct {
	mu   sync.Mutex
	data map[string]string
}

// GOOD: zero-value variable
var mu sync.Mutex
