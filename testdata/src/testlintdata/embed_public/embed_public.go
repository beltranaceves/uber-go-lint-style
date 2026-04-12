package embed_public

// BAD: exported struct embedding exported type
type ConcreteList struct {
	*AbstractList // want "avoid embedding exported type 'AbstractList' in exported struct 'ConcreteList'"
}

// AbstractList is an exported implementation type (simulates library type)
type AbstractList struct{}

// GOOD: unexported struct embedding exported type is allowed
type concreteList struct {
	*AbstractList
}

// GOOD: exported struct with non-exported embedded type is allowed
type impl struct{}

type PublicWrapper struct {
	*impl
}
