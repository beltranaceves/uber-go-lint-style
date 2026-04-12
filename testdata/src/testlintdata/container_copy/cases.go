package container_copy

// Test intent: additional BAD and GOOD examples for container_copy rule

// BAD: assigning parameter slice directly to receiver field
func (d *Driver) MultiAssign(trips []Trip, other []Trip) {
	d.trips = trips // want "copy slice or map when storing or returning to avoid sharing underlying data"
	d.trips = other // want "copy slice or map when storing or returning to avoid sharing underlying data"
}

// BAD: map assigned directly
type MStats struct {
	m map[string]int
}

func (s *MStats) SetCounters(c map[string]int) {
	s.m = c // want "copy slice or map when storing or returning to avoid sharing underlying data"
}

// GOOD: copy before assigning
func (d *Driver) SafeSet(trips []Trip) {
	d.trips = make([]Trip, len(trips))
	copy(d.trips, trips)
}

// GOOD: another safe example
func (d *Driver) AnotherSafe(trips []Trip) {
	tmp := make([]Trip, len(trips))
	copy(tmp, trips)
	d.trips = tmp
}
