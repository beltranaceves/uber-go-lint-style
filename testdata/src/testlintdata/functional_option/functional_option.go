package functional_option

// Minimal types to make the examples compile in analysistest.
type Connection struct{}

type options struct{ cache bool }

type Option interface{ apply(*options) }

type cacheOption bool

func (c cacheOption) apply(o *options) { o.cache = bool(c) }

func WithCache(b bool) Option { return cacheOption(b) }

// BAD: exported function with three parameters should use functional options
func Open(addr string, cache bool, logger *int) (*Connection, error) { // want "exported function has 3 or more parameters; consider using the functional options pattern for optional arguments"
	return nil, nil
}

// GOOD: functional options used instead of multiple parameters
func OpenWithOpts(addr string, opts ...Option) (*Connection, error) {
	return nil, nil
}
