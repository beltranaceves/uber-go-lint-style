package function_order

// Function ordering violations

func (s *something) Stop() {}

type something struct{} // want "types/const/var declarations must appear before functions"

func newSomething() *something { return &something{} } // want "constructor newSomething should appear immediately after type something"

type A struct{}
type B struct{}

func (a *A) One()   {}
func (b *B) First() {}
func (a *A) Two()   {} // want "methods of receiver 'A' must be contiguous"

type C struct{}

func (c *C) unexported() {}
func (c *C) Exported()   {} // want "exported method 'Exported' should appear before unexported methods for receiver 'C'"

type D struct{}

func (d *D) Callee() {}
func (d *D) Caller() { d.Callee() } // want "method 'Caller' calls 'Callee' but appears after it; declare 'Caller' before 'Callee'"
