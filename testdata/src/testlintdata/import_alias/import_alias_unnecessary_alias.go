package import_alias

import (
	runtimetrace "runtime/trace" // want "unnecessary import alias 'runtimetrace' for package 'trace'; remove the alias"
)

var _ = runtimetrace.Start
