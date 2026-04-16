package param_naked

// Example function with two boolean parameters
func printInfo(name string, isLocal, done bool) {}

func callerBad() {
	// BAD: boolean literals are naked and should be commented or replaced
	b := false
	printInfo("foo", true, b) // want "avoid naked boolean parameter \"isLocal\"; add an inline comment at callsite or use a named type"
	printInfo("foo", b, true) // want "avoid naked boolean parameter \"done\"; add an inline comment at callsite or use a named type"
}

func callerGood() {
	// GOOD: boolean literals annotated with comments
	printInfo("foo", true /* isLocal */, true /* done */)
}
