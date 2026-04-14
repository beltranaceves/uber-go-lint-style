package import_alias_missing

import (
	"example.com/trace/v2" // want "import path \"example.com/trace/v2\" package name \"trace\" does not match last path element \"v2\"; add an explicit alias \"trace\""
)

var _ = trace.Foo
