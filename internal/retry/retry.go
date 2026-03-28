package retry

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxAttempts int           // Maximum number of attempts
	InitialWait time.Duration // Initial wait between retries
	MaxWait     time.Duration // Maximum wait between retries
	Multiplier  float64       // Backoff multiplier for exponential backoff
}

// DefaultConfig returns default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     5 * time.Second,
		Multiplier:  2.0,
	}
}

// Do executes a function with exponential backoff retry logic
// Returns error if all attempts fail
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Try the function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't wait after the last failed attempt
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// Calculate wait time with exponential backoff
		waitTime := calculateBackoff(attempt, cfg.InitialWait, cfg.MaxWait, cfg.Multiplier)

		// Wait with context cancellation support
		select {
		case <-time.After(waitTime):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

// calculateBackoff calculates exponential backoff with jitter
func calculateBackoff(attempt int, initial, max time.Duration, multiplier float64) time.Duration {
	// Exponential backoff: initial * (multiplier ^ attempt)
	backoff := time.Duration(float64(initial) * math.Pow(multiplier, float64(attempt)))

	// Cap at max wait
	if backoff > max {
		backoff = max
	}

	return backoff
}
