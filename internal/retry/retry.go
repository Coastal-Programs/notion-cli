// Package retry provides exponential backoff with jitter for retrying
// operations that may fail transiently.
package retry

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"time"
)

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Jitter     bool
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults:
// 3 retries, 1s base delay, 30s max delay, jitter enabled.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Jitter:     true,
	}
}

// RetryableError represents an error that may be retried, optionally
// including an HTTP status code and a Retry-After duration.
type RetryableError struct {
	Err        error
	StatusCode int
	RetryAfter time.Duration
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// Do executes fn with retry logic according to cfg.
// It returns nil on success, or the last error if all retries are exhausted.
// It respects context cancellation between retries.
func Do(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Check if the error is retryable
		if !isRetryableErr(lastErr) {
			return lastErr
		}

		// Don't sleep after the last attempt
		if attempt == cfg.MaxRetries {
			break
		}

		// Calculate delay
		delay := CalculateDelay(attempt, cfg)

		// Check for RetryAfter hint
		var retryErr *RetryableError
		if errors.As(lastErr, &retryErr) && retryErr.RetryAfter > 0 {
			delay = retryErr.RetryAfter
		}

		// Wait or cancel
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}

// IsRetryable returns true if the given HTTP status code indicates a
// retryable error: 408, 429, 500, 502, 503, 504.
func IsRetryable(statusCode int) bool {
	switch statusCode {
	case 408, 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

// CalculateDelay computes the backoff delay for the given attempt number.
// The delay is min(baseDelay * 2^attempt, maxDelay), with optional equal
// jitter that guarantees at least 50% of the calculated delay.
func CalculateDelay(attempt int, cfg RetryConfig) time.Duration {
	delay := float64(cfg.BaseDelay) * math.Pow(2, float64(attempt))
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}

	if cfg.Jitter {
		half := delay / 2
		delay = half + rand.Float64()*half
	}

	return time.Duration(delay)
}

// isRetryableErr checks whether an error should be retried.
// It returns true for *RetryableError instances and false for all others.
func isRetryableErr(err error) bool {
	var retryErr *RetryableError
	return errors.As(err, &retryErr)
}
