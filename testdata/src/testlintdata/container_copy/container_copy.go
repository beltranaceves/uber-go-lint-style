package container_copy

import "sync"

type Trip struct{}

type Driver struct {
	trips []Trip
}

func (d *Driver) SetTrips(trips []Trip) {
	d.trips = trips // want "copy slice or map when storing or returning to avoid sharing underlying data"
}

type Stats struct {
	mu       sync.Mutex
	counters map[string]int
}

// Snapshot returns the current stats.
func (s *Stats) Snapshot() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.counters // want "copy slice or map when storing or returning to avoid sharing underlying data"
}

func (d *Driver) SetTripsGood(trips []Trip) {
	d.trips = make([]Trip, len(trips))
	copy(d.trips, trips)
}

func (s *Stats) SnapshotGood() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]int, len(s.counters))
	for k, v := range s.counters {
		result[k] = v
	}
	return result
}
