package structpointer

type T struct{ Name string }

// BAD: using new(T) for struct initialization
func bad1() {
    sval := T{Name: "foo"}
    _ = sval

    // inconsistent
    sptr := new(T) // want "use &T instead of new T when initializing struct references"
    sptr.Name = "bar"
    _ = sptr
}

// BAD: another named struct
func bad2() {
    type S struct{ A int }
    _ = new(S) // want "use &T instead of new T when initializing struct references"
}

// GOOD: use composite literal with &
func good1() {
    sval := T{Name: "foo"}
    sptr := &T{Name: "bar"}
    _ = sval
    _ = sptr
}

// GOOD: new for non-struct types is fine
func good2() {
    i := new(int)
    _ = i
}
