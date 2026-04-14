package import_group

// GOOD: stdlib then blank line then third-party
import (
	"bytes"

	"go.uber.org/atomic"
)

func _() { _ = bytes.NewBuffer(nil); _ = atomic.Value{} }
