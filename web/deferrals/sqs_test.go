package deferrals

import (
	"fmt"
	"testing"
	"time"
)

func TestDefer(t *testing.T) {
	sqs := NewSQS(awsCoordinates())
	relay := make(chan Deferral)
	svc := NewSQSDeferralService(sqs, relay)

	if err := svc.CreateQueue("testq"); err != nil {
		t.Error(err)
	}

	if err := svc.Defer("proj", "feature/foo"); err != nil {
		t.Error(err)
	}
	fmt.Println("Defer")

	for j := 0; j < 10; j++ {
		fmt.Println("Resubmit")
		svc.Resubmit()
		time.Sleep(2 * time.Second)
	}

}
