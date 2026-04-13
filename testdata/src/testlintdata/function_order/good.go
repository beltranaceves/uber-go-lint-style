package function_order

type somethingGood struct{}
type AGood struct{}
type CGood struct{}
type DGood struct{}

func newSomethingGood() *somethingGood { return &somethingGood{} }

func (s *somethingGood) Cost() int { return s.calc() }
func (s *somethingGood) calc() int { return 0 }

func (a *AGood) One() {}
func (a *AGood) Two() {}

func (c *CGood) Exported()   {}
func (c *CGood) unexported() {}
func (d *DGood) Caller()     { d.Callee() }
func (d *DGood) Callee()     {}
