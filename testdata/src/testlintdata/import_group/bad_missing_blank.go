package import_group

// BAD: missing blank line between std and other
import (
	"fmt"
	"github.com/foo/bar" // want "add blank line between standard library and other imports"
)

func _() { fmt.Println(); _ = bar.X }
