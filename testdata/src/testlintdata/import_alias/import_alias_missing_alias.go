package import_alias

import "example.com/client-go" // want "import path \"example.com/client-go\" package name \"client\" does not match last path element \"client-go\"; add an explicit alias \"client\""

var _ = client.Hello
