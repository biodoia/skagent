package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultConfig returns sensible defaults for API calls
func DefaultConfig() Config {
	return Config{
		MaxRetries:  3,
		InitialWait: 1 * time.Second,
		MaxWait:     30 * time.Second,
		Multiplier:  2.0,
	}
}

// Retryable errors
var (
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
)

// IsRetryable determines if an error should trigger a retry
type IsRetryable func(error) bool

// DefaultIsRetryable returns true for temporary/network errors
func DefaultIsRetryable(err error) bool {
	if err == nil {
		return false
	}
	// Retry on context deadline (request timeout, not user cancel)
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// Don't retry on context canceled (user action)
	if errors.Is(err, context.Canceled) {
		return false
	}
	// Retry on network/temporary errors (simple heuristic)
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"503",
		"502",
		"429", // rate limit
	}
	for _, pattern := range retryablePatterns {
		if containsInsensitive(errStr, pattern) {
			return true
		}
	}
	return false
}

// Do executes fn with retry logic
func Do(ctx context.Context, cfg Config, isRetryable IsRetryable, fn func() error) error {
	if isRetryable == nil {
		isRetryable = DefaultIsRetryable
	}

	var lastErr error
	wait := cfg.InitialWait

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return ctx.Err()
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Don't retry if not retryable
		if !isRetryable(lastErr) {
			return lastErr
		}

		// Don't wait after the last attempt
		if attempt == cfg.MaxRetries {
			break
		}

		// Wait with exponential backoff
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
			// Calculate next wait time
			wait = time.Duration(float64(wait) * cfg.Multiplier)
			if wait > cfg.MaxWait {
				wait = cfg.MaxWait
			}
		}
	}

	return errors.Join(ErrMaxRetriesExceeded, lastErr)
}

// DoWithResult executes fn and returns result with retry logic
func DoWithResult[T any](ctx context.Context, cfg Config, isRetryable IsRetryable, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	err := Do(ctx, cfg, isRetryable, func() error {
		var fnErr error
		result, fnErr = fn()
		lastErr = fnErr
		return fnErr
	})

	if err != nil {
		return result, err
	}
	return result, lastErr
}

// ExponentialBackoff calculates delay for a given attempt
func ExponentialBackoff(attempt int, initial, max time.Duration, multiplier float64) time.Duration {
	delay := time.Duration(float64(initial) * math.Pow(multiplier, float64(attempt)))
	if delay > max {
		return max
	}
	return delay
}

func containsInsensitive(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 len(substr) > 0 && 
		 findInsensitive(s, substr) >= 0)
}

func findInsensitive(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(s) < len(substr) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc, tc := s[i+j], substr[j]
			// Simple ASCII case-insensitive compare
			if sc != tc && sc != tc+32 && sc != tc-32 {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
