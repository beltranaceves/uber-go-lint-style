package import_group

// GOOD: single third-party import
import "github.com/pkg/errors"

func _() { _ = errors.New("x") }
