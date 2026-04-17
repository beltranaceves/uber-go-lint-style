package printf_const

import "fmt"

var msg = "unexpected values %v, %v\n"

func Bad() {
	fmt.Printf(msg, 1, 2) // want "format string should be a const value"
}

const cmsg = "unexpected values %v, %v\n"

func GoodConst() {
	fmt.Printf(cmsg, 1, 2)
}

func GoodLiteral() {
	fmt.Printf("ok %v\n", 1)
}
