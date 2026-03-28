package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoSuccess(t *testing.T) {
	callCount := 0
	cfg := DefaultConfig()
	cfg.MaxAttempts = 3

	err := Do(context.Background(), cfg, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestDoRetrySuccess(t *testing.T) {
	callCount := 0
	cfg := DefaultConfig()
	cfg.MaxAttempts = 3

	err := Do(context.Background(), cfg, func() error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestDoAllAttemptsFail(t *testing.T) {
	callCount := 0
	cfg := DefaultConfig()
	cfg.MaxAttempts = 3

	err := Do(context.Background(), cfg, func() error {
		callCount++
		return errors.New("permanent error")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestDoContextCancellation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxAttempts = 5
	cfg.InitialWait = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	callCount := 0
	err := Do(ctx, cfg, func() error {
		callCount++
		if callCount == 1 {
			cancel() // Cancel after first attempt
		}
		return errors.New("error")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}

	if callCount != 1 {
		t.Errorf("expected 1 call before cancellation, got %d", callCount)
	}
}
