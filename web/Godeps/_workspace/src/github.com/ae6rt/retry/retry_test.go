package retry_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ae6rt/retry"
)

func TestOKNoTimeout(t *testing.T) {
	r := retry.New(0*time.Second, 3, retry.DefaultBackoffFunc)
	err := r.Try(func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
}

func TestOKWithTimeout(t *testing.T) {
	r := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)
	err := r.Try(func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
}

func TestRetryExceeded(t *testing.T) {
	r := retry.New(0*time.Second, 3, retry.DefaultBackoffFunc)
	tries := 0
	err := r.Try(func() error {
		tries += 1
		return errors.New("")
	})
	if err == nil {
		t.Fatalf("Expecting error\n")
	}
	if tries != 3 {
		t.Fatalf("Expecting 3 but got %d\n", tries)
	}
}

func TestTimeout(t *testing.T) {
	r := retry.New(500*time.Millisecond, 1, retry.DefaultBackoffFunc)
	err := r.Try(func() error {
		time.Sleep(1000 * time.Millisecond)
		return nil
	})
	if err == nil {
		t.Fatalf("Expected error\n")
	}
	fmt.Println(err)
	if !retry.IsTimeout(err) {
		t.Fatalf("Expected retry.timeoutError\n")
	}
}
