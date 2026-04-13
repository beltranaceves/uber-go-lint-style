package enum_start

type Operation int

const (
	Add Operation = iota // want "enum 'Operation' starts at 0; prefer starting at 1"
	Subtract
)

type Operation2 int

const (
	Ok Operation2 = iota + 1
	Cancel
)

const (
	A = iota
	B
)
