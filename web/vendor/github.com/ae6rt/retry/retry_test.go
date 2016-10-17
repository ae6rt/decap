package retry_test

import (
	"errors"
	"testing"

	"github.com/ae6rt/retry"
)

func TestRetryExceeded(t *testing.T) {
	r := retry.New(3, retry.DefaultBackoffFunc)
	tries := 0
	err := r.Try(func() error {
		tries += 1
		return errors.New("woops")
	})
	if err == nil {
		t.Fatalf("Expecting error\n")
	}
	if tries != 3 {
		t.Fatalf("Expecting 3 but got %d\n", tries)
	}
}
