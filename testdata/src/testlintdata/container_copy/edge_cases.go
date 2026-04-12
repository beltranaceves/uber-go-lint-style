package container_copy

// Test intent: edge cases for container_copy rule

// BAD: assigning parameter to another instance's field should NOT trigger
func (d *Driver) AssignToOther(o *Driver, trips []Trip) {
	o.trips = trips
}

// BAD: value receiver assignment (analyzed the same way; tool will still flag)
func (d Driver) ValueReceiverSet(trips []Trip) {
	d.trips = trips // want "copy slice or map when storing or returning to avoid sharing underlying data"
}

// GOOD: assigning non-parameter (local variable) shouldn't trigger
func (d *Driver) AssignLocal(trips []Trip) {
	tmp := trips
	d.trips = tmp
}
