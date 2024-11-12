package util

import (
	"log"
	"time"
)

func RetryWithBackoff(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		log.Printf("Attempt %d failed: %v. Retrying in %s...", i+1, err, delay)
		time.Sleep(delay)
		// Exponential backoff
		delay *= 2
	}
	return err
}
