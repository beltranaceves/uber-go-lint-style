package interface_receiver

type T struct{ a int }

func (t T) Mv() int { return t.a }
func (t *T) Mp()    {}

func examples() {
	t := new(T)

	f := t.Mp // want "taking a method value with a pointer receiver captures the receiver; subsequent mutations to the pointee will not affect the stored receiver"
	_ = f

	g := t.Mv // want "taking a method value captures the receiver by value; subsequent mutations to the original value will not affect the stored receiver"
	_ = g

	// method call should not be flagged
	t.Mp()
}
