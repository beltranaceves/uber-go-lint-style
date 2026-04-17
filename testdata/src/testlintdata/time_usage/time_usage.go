package time_usage

import "time"

// BAD: parameter named like a time instant but using int
func badParam(now int) { // want "prefer time.Time for instants and time.Duration for durations"
	_ = now
}

// GOOD: use time.Time
func goodParam(now time.Time) {
	_ = now
}

// BAD: Sleep with numeric literal
func badSleep() {
	time.Sleep(10) // want "use time.Duration with time.Sleep"
}

// GOOD: use time.Duration
func goodSleep() {
	time.Sleep(10 * time.Second)
}
