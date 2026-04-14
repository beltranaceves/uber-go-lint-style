package import_alias

import (
	"fmt"

	client "example.com/client-go"
)

func _() {
	// Use imports to avoid unused-import errors in tests
	_ = fmt.Println
	_ = client.Hello
}
