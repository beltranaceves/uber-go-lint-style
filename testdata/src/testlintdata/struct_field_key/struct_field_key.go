package struct_field_key

type User struct {
	FirstName string
	LastName  string
	Admin     bool
}

func GoodExample() {
	_ = User{
		FirstName: "John",
		LastName:  "Doe",
		Admin:     true,
	}
}

func BadExample() {
	_ = User{"John", "Doe", true} // want "use field names when initializing structs; specify fields like `Field: value`"
}

func TestTableAllowed() {
	tests := []struct {
		op   int
		want string
	}{
		{1, "one"},
	}
	_ = tests
}

func TestTableTooLarge() {
	tests := []struct {
		a int
		b int
		c int
		d int
	}{
		{1, 2, 3, 4}, // want "use field names when initializing structs; specify fields like `Field: value`"
	}
	_ = tests
}
