package slice_nil

func ReturnEmpty() []int {
	if true {
		return []int{} // want "prefer returning nil for zero-length slices"
	}
	return nil
}

func ReturnMakeEmpty() []string {
	return make([]string, 0) // want "prefer returning nil for zero-length slices"
}

func ReturnNil() []int {
	return nil
}

func VarInitBad() {
	nums := []int{} // want "prefer nil slice"
	_ = nums
}

func VarInitGood() {
	var nums []int
	_ = nums
}

func CheckNilBad(s []string) bool {
	return s == nil // want "use len"
}

func CheckLenGood(s []string) bool {
	return len(s) == 0
}
