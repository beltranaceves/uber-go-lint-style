// Auto-generated test case for error-wrap rule
// Negative = should PASS lint (Good code)

package testdata

import (
	"database/sql"
	"fmt"
)

// Example: Wrapping errors with context (GOOD)
func QueryUser(db *sql.DB, id string) error {
	rows, err := db.Query("SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("query failed: %w", err) // Wrapped with context
	}
	defer rows.Close()

	if err := rows.Scan(id); err != nil {
		return fmt.Errorf("scan failed: %w", err) // Wrapped with context
	}

	return nil
}
