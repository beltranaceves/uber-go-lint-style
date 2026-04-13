package error_once

import (
	"fmt"
	"log"
)

func getUser() (string, error)  { return "", nil }
func emitMetrics() (int, error) { return 0, nil }

func exampleBad() error {
	u, err := getUser()
	if err != nil {
		// BAD: Log and return
		log.Printf("could not get user: %v", err)
		return err // want "handle error only once: avoid logging the error and then returning it"
	}
	_ = u
	return nil
}

func exampleGoodWrap() error {
	u, err := getUser()
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	_ = u
	return nil
}

func exampleGoodLogAndRecover() {
	_, err := emitMetrics()
	if err != nil {
		log.Printf("could not emit metrics: %v", err)
		// degraded behavior; no return
	}
}

// Structured logger example
type Logger struct{}

func (Logger) Errorf(format string, args ...any) {}

var logger Logger

func exampleBadStructured() error {
	u, err := getUser()
	if err != nil {
		logger.Errorf("could not get user: %v", err)
		return err // want "handle error only once: avoid logging the error and then returning it"
	}
	_ = u
	return nil
}

func exampleGoodWrapped() error {
	u, err := getUser()
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	_ = u
	return nil
}
