package retry

import (
	"fmt"
	"log"
	"os"
	"time"
)

var Log *log.Logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

var DefaultBackoffFunc = func(attempts int) {
	time.Sleep((1 << uint(attempts)) * time.Second)
}

func New(maxAttempts int, backoffFunc func(int)) Retry {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return Retry{maxAttempts: maxAttempts, backoffFunc: backoffFunc}
}

type Retry struct {
	maxAttempts int
	backoffFunc func(int)
}

func (r Retry) Try(work func() error) error {
	var err error
	for i := 0; ; i++ {
		err = work()
		if err == nil {
			return nil
		}

		if i >= (r.maxAttempts - 1) {
			break
		}

		r.backoffFunc(i)
		Log.Println("retrying...")
	}
	return fmt.Errorf("after %d attempts, last error: %s", r.maxAttempts, err)
}
