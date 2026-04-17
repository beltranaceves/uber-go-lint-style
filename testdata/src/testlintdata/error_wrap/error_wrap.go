package error_wrap

import "fmt"

func Bad() error {
	err := fmt.Errorf("the error")
	return fmt.Errorf("failed to create new store: %w", err) // want "avoid 'failed to' prefix in error messages; use concise context, e.g. new store: %w"
}

func Good() error {
	err := fmt.Errorf("the error")
	return fmt.Errorf("new store: %w", err)
}
