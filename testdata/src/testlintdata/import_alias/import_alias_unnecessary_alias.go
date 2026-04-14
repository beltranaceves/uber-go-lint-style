package import_alias

import (
	fmtalias "fmt"               // want "unnecessary import alias 'fmtalias' for package 'fmt'; remove the alias"
	runtimetrace "runtime/trace" // want "unnecessary import alias 'runtimetrace' for package 'trace'; remove the alias"
)

var _ = runtimetrace.Start
var _ = fmtalias.Sprint
