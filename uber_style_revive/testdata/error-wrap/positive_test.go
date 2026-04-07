// Auto-generated test case for error-wrap rule
// Positive = should FAIL lint (Bad code)

package testdata

import "database/sql"

// Example: Not wrapping errors (BAD)
func QueryUser(db *sql.DB, id string) error {
	rows, err := db.Query("SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return err // Should be wrapped with context
	}
	defer rows.Close()

	return rows.Scan(id) // Bare error return without wrapping
}
