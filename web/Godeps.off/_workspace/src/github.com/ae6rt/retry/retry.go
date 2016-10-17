package retry

import (
	"time"
)

var DefaultBackoffFunc = func(attempts uint) {
	if attempts == 0 {
		return
	}
	time.Sleep((1 << attempts) * time.Millisecond)
}

func New(timeout time.Duration, maxAttempts uint, backoffFunc func(uint)) Retry {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return Retry{timeout: timeout, maxAttempts: maxAttempts, backoffFunc: backoffFunc}
}

type timeoutError struct {
	error
}

func (t timeoutError) Error() string {
	return "retry.timeout"
}

type Retry struct {
	timeout     time.Duration
	maxAttempts uint
	backoffFunc func(uint)
}

func (r Retry) Try(work func() error) error {
	doneChan := make(chan struct{})
	errorChan := make(chan error)
	var attempts uint = 0

	var expired <-chan time.Time
	if r.timeout > 0 {
		timer := time.NewTimer(r.timeout)
		expired = timer.C
		defer timer.Stop()
	}

	for {
		go func() {
			r.backoffFunc(attempts)
			attempts += 1
			if err := work(); err != nil {
				errorChan <- err
			} else {
				doneChan <- struct{}{}
			}
		}()

		select {
		case <-doneChan:
			return nil
		case err := <-errorChan:
			if attempts == r.maxAttempts {
				return err
			}
		case <-expired:
			return timeoutError{}
		}
	}
}

func IsTimeout(err error) bool {
	if _, ok := err.(timeoutError); ok {
		return true
	}
	return false
}
