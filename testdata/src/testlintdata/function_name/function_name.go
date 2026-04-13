package function_name

// BAD: contains underscore
func bad_name() {} // want "should use MixedCaps"

// GOOD: MixedCaps
func GoodName() {}

// GOOD: lowerCamel is acceptable
func goodName() {}

// BAD: lowercase-only name
func goodnameonly() {} // want "function name is lowercase-only"

// BAD: internal helper with underscore
func helper_do_thing() {} // want "should use MixedCaps"
