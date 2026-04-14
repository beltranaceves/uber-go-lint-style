package import_group

// BAD: three groups (stdlib, third-party, stdlib)
import (
	"fmt"
	"github.com/x/y"
	"net/http" // want "imports must be grouped: standard library first, then third-party imports"
)

func _() { _ = fmt.Print; _ = http.DefaultClient; _ = y.Y }
