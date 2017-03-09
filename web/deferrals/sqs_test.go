package deferrals

import (
	"fmt"
	"sync"
	"testing"
)

func TestDefer(t *testing.T) {
	var wg sync.WaitGroup

	relay := make(chan Deferral)
	go func() {
		msg := <-relay
		fmt.Printf("@@@ TestDefer() received channel deferral %+v\n", msg)
		wg.Done()
	}()

	sqs := NewSQS(awsCoordinates())
	svc := NewSQSDeferralService(sqs, relay)

	if err := svc.CreateQueue("testq"); err != nil {
		t.Error(err)
	}

	if err := svc.Defer("proj", "feature/foo"); err != nil {
		t.Error(err)
	}

	wg.Add(1)
	fmt.Println("Defer")

	fmt.Println("Resubmit")
	svc.Resubmit()

	wg.Wait()
}
