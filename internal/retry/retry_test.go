package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", cfg.MaxRetries)
	}
	if cfg.BaseDelay != 1*time.Second {
		t.Errorf("expected BaseDelay=1s, got %v", cfg.BaseDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay=30s, got %v", cfg.MaxDelay)
	}
	if !cfg.Jitter {
		t.Error("expected Jitter=true")
	}
}

func TestDoSuccess(t *testing.T) {
	cfg := DefaultRetryConfig()
	calls := 0

	err := Do(context.Background(), cfg, func() error {
		calls++
		return nil
	})

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestDoNonRetryableError(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.BaseDelay = 1 * time.Millisecond
	calls := 0

	nonRetryable := errors.New("bad request")

	err := Do(context.Background(), cfg, func() error {
		calls++
		return nonRetryable
	})

	if err != nonRetryable {
		t.Errorf("expected original error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retry for non-retryable), got %d", calls)
	}
}

func TestDoRetryableEventualSuccess(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Jitter:     false,
	}
	calls := 0

	err := Do(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return &RetryableError{
				Err:        fmt.Errorf("server error"),
				StatusCode: 503,
			}
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected nil error after retries, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDoRetryableExhausted(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 2,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Jitter:     false,
	}
	calls := 0

	err := Do(context.Background(), cfg, func() error {
		calls++
		return &RetryableError{
			Err:        fmt.Errorf("always failing"),
			StatusCode: 500,
		}
	})

	if err == nil {
		t.Error("expected error after exhausting retries")
	}
	// initial + 2 retries = 3 calls
	if calls != 3 {
		t.Errorf("expected 3 calls (1 initial + 2 retries), got %d", calls)
	}
}

func TestDoContextCancellation(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 5,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		Jitter:     false,
	}

	ctx, cancel := context.WithCancel(context.Background())
	calls := 0

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, cfg, func() error {
		calls++
		return &RetryableError{
			Err:        fmt.Errorf("failing"),
			StatusCode: 503,
		}
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDoRetryAfterHint(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 2,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		Jitter:     false,
	}
	calls := 0
	start := time.Now()

	err := Do(context.Background(), cfg, func() error {
		calls++
		if calls == 1 {
			return &RetryableError{
				Err:        fmt.Errorf("rate limited"),
				StatusCode: 429,
				RetryAfter: 50 * time.Millisecond,
			}
		}
		return nil
	})

	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("expected at least 40ms delay from RetryAfter, got %v", elapsed)
	}
}

func TestIsRetryable(t *testing.T) {
	retryableCodes := []int{408, 429, 500, 502, 503, 504}
	nonRetryableCodes := []int{200, 201, 400, 401, 403, 404, 405, 409, 422}

	for _, code := range retryableCodes {
		if !IsRetryable(code) {
			t.Errorf("expected status %d to be retryable", code)
		}
	}

	for _, code := range nonRetryableCodes {
		if IsRetryable(code) {
			t.Errorf("expected status %d to NOT be retryable", code)
		}
	}
}

func TestCalculateDelayNoJitter(t *testing.T) {
	cfg := RetryConfig{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  10 * time.Second,
		Jitter:    false,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 100 * time.Millisecond},  // 100ms * 2^0 = 100ms
		{1, 200 * time.Millisecond},  // 100ms * 2^1 = 200ms
		{2, 400 * time.Millisecond},  // 100ms * 2^2 = 400ms
		{3, 800 * time.Millisecond},  // 100ms * 2^3 = 800ms
		{4, 1600 * time.Millisecond}, // 100ms * 2^4 = 1600ms
	}

	for _, tt := range tests {
		delay := CalculateDelay(tt.attempt, cfg)
		if delay != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, delay)
		}
	}
}

func TestCalculateDelayMaxCap(t *testing.T) {
	cfg := RetryConfig{
		BaseDelay: 1 * time.Second,
		MaxDelay:  5 * time.Second,
		Jitter:    false,
	}

	// 1s * 2^10 = 1024s, capped at 5s
	delay := CalculateDelay(10, cfg)
	if delay != 5*time.Second {
		t.Errorf("expected delay capped at 5s, got %v", delay)
	}
}

func TestCalculateDelayWithJitter(t *testing.T) {
	cfg := RetryConfig{
		BaseDelay: 1 * time.Second,
		MaxDelay:  30 * time.Second,
		Jitter:    true,
	}

	// Run multiple times to verify randomness
	delays := make(map[time.Duration]bool)
	for i := 0; i < 20; i++ {
		delay := CalculateDelay(2, cfg)
		delays[delay] = true

		// With jitter, delay should be 0 to 4s (baseDelay * 2^2 = 4s * rand[0,1))
		if delay < 0 || delay > 4*time.Second {
			t.Errorf("jittered delay out of range: %v", delay)
		}
	}

	// With 20 attempts, we should see some variation
	if len(delays) < 2 {
		t.Error("expected jitter to produce varying delays")
	}
}

func TestRetryableErrorError(t *testing.T) {
	err := &RetryableError{
		Err:        fmt.Errorf("server error"),
		StatusCode: 500,
	}

	if err.Error() != "server error" {
		t.Errorf("expected 'server error', got %q", err.Error())
	}
}

func TestRetryableErrorUnwrap(t *testing.T) {
	inner := fmt.Errorf("inner error")
	err := &RetryableError{
		Err:        inner,
		StatusCode: 503,
	}

	if !errors.Is(err, inner) {
		t.Error("expected Unwrap to expose inner error")
	}
}

func TestDoZeroRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 0,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Jitter:     false,
	}
	calls := 0

	err := Do(context.Background(), cfg, func() error {
		calls++
		return &RetryableError{
			Err:        fmt.Errorf("failing"),
			StatusCode: 500,
		}
	})

	if err == nil {
		t.Error("expected error with 0 retries")
	}
	if calls != 1 {
		t.Errorf("expected 1 call with 0 retries, got %d", calls)
	}
}

func TestDoMixedErrors(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 5,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Jitter:     false,
	}
	calls := 0

	err := Do(context.Background(), cfg, func() error {
		calls++
		if calls <= 2 {
			return &RetryableError{
				Err:        fmt.Errorf("transient"),
				StatusCode: 503,
			}
		}
		// Third call returns non-retryable error
		return errors.New("permanent failure")
	})

	if err == nil || err.Error() != "permanent failure" {
		t.Errorf("expected 'permanent failure', got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestIsRetryableErr(t *testing.T) {
	retryable := &RetryableError{Err: errors.New("test"), StatusCode: 500}
	nonRetryable := errors.New("regular error")

	if !isRetryableErr(retryable) {
		t.Error("expected RetryableError to be retryable")
	}
	if isRetryableErr(nonRetryable) {
		t.Error("expected regular error to not be retryable")
	}
}

func TestDoContextDeadline(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 10,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Jitter:     false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := Do(ctx, cfg, func() error {
		return &RetryableError{
			Err:        fmt.Errorf("timeout"),
			StatusCode: 503,
		}
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}
