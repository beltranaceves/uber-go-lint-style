package container_capacity

import "os"

// BAD: map created without capacity then populated in a loop
func badMap() {
	files, _ := os.ReadDir("./files")

	m := make(map[string]os.DirEntry) // want "preallocate map capacity when populating in a loop"
	for _, f := range files {
		m[f.Name()] = f
	}
	_ = m
}

// GOOD: map created with capacity
func goodMap() {
	files, _ := os.ReadDir("./files")

	m := make(map[string]os.DirEntry, len(files))
	for _, f := range files {
		m[f.Name()] = f
	}
	_ = m
}

// BAD: slice created without capacity then appended
func badSlice(b int, size int) {
	for n := 0; n < b; n++ {
		data := make([]int, 0) // want "preallocate slice capacity when appending in a loop"
		for k := 0; k < size; k++ {
			data = append(data, k)
		}
		_ = data
	}
}

// GOOD: slice created with capacity
func goodSlice(b int, size int) {
	for n := 0; n < b; n++ {
		data := make([]int, 0, size)
		for k := 0; k < size; k++ {
			data = append(data, k)
		}
		_ = data
	}
}
