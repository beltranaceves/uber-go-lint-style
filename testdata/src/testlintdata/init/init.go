package initpkg

// -----------------
// Example 1: Default Foo initialization
// -----------------

type Foo struct {
	A int
}

var _defaultFoo Foo

// BAD: init assigns to package var — should be flagged
func init() { // want "avoid init functions; prefer explicit initialization in main or helper functions"
	_defaultFoo = Foo{A: 1}
}

// GOOD: direct var initialization
var _defaultFooGood = Foo{A: 1}

// GOOD: helper-based initialization
var _defaultFooBetter = defaultFoo()

func defaultFoo() Foo {
	return Foo{A: 1}
}

// -----------------
// Example 2: Config loading
// -----------------

type Config struct {
	Path string
}

var _config Config

// BAD: init performs I/O and environment access — should be flagged
func init() { // want "avoid init functions; prefer explicit initialization in main or helper functions"
	// emulate accessing working directory and reading file
	_config = Config{Path: "/tmp/config.yaml"}
}

// GOOD: explicit loader function that can be called from main
func loadConfig() Config {
	// in real code: os.Getwd(), os.ReadFile, yaml.Unmarshal, etc.
	return Config{Path: "/tmp/config.yaml"}
}
