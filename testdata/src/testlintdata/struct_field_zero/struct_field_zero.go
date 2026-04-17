package struct_field_zero

type User struct {
	FirstName  string
	LastName   string
	MiddleName string
	Admin      bool
	Ptr        *int
}

func GoodExample() {
	_ = User{
		FirstName: "John",
		LastName:  "Doe",
	}
}

func BadExample() {
	_ = User{
		FirstName:  "John",
		MiddleName: "",    // want "omit zero-valued field \"MiddleName\" from struct literal; let Go set the zero value"
		Admin:      false, // want "omit zero-valued field \"Admin\" from struct literal; let Go set the zero value"
		Ptr:        nil,   // want "omit zero-valued field \"Ptr\" from struct literal; let Go set the zero value"
	}
}

func TestTableAllowed() {
	tests := []struct {
		give string
		want int
	}{
		{give: "0", want: 0},
	}
	_ = tests
}
