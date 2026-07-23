package database

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitReadySucceedsAfterTransientFailures(t *testing.T) {
	attempts := 0
	err := waitReady(context.Background(), 10, time.Millisecond, func(context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("the database system is starting up")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("waitReady() = %v, want nil", err)
	}
	if attempts != 3 {
		t.Fatalf("ping called %d times, want 3", attempts)
	}
}

func TestWaitReadyGivesUpAfterAllAttempts(t *testing.T) {
	wantErr := errors.New("the database system is starting up")
	attempts := 0
	err := waitReady(context.Background(), 4, time.Millisecond, func(context.Context) error {
		attempts++
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("waitReady() = %v, want wrapping %v", err, wantErr)
	}
	if attempts != 4 {
		t.Fatalf("ping called %d times, want 4", attempts)
	}
}

func TestWaitReadyStopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	err := waitReady(ctx, 10, 50*time.Millisecond, func(context.Context) error {
		attempts++
		if attempts == 1 {
			cancel()
		}
		return errors.New("the database system is starting up")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("waitReady() = %v, want context.Canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("ping called %d times, want 1", attempts)
	}
}
