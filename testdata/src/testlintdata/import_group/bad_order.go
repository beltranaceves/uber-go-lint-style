package import_group

// BAD: third-party imports before stdlib
import (
	"fmt"
	"github.com/foo/bar" // want "add blank line between standard library and other imports"
)

func _() { _ = fmt.Print; _ = bar.X }
