package line_length

// This comment is short and fine.

func Good() {
	// ok
}

var longStr = "This string literal is intentionally very long to exceed the ninety-nine character soft limit to trigger the linter rule." // want "line exceeds recommended 99 character limit"
