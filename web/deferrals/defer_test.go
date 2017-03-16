package deferrals

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
)

type DeferMock struct {
	MockSQS
	queueURL         string
	createQueueInput *sqs.CreateQueueInput
	messageInput     *sqs.SendMessageInput
}

func (t *DeferMock) CreateQueue(f *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	t.createQueueInput = f
	return &sqs.CreateQueueOutput{QueueUrl: &t.queueURL}, nil
}

func (t *DeferMock) SendMessage(f *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	t.messageInput = f
	return &sqs.SendMessageOutput{}, nil
}

func TestDefer(t *testing.T) {
	var tests = []struct {
		wantQueueName  string
		wantQueueURL   string
		wantProjectKey string
		wantBranch     string
	}{
		{
			wantQueueName:  "foo",
			wantQueueURL:   "http://example.com",
			wantProjectKey: "proj",
			wantBranch:     "issue/1",
		},
	}

	for testNumber, test := range tests {
		mockSQS := &DeferMock{queueURL: test.wantQueueURL}

		deferralService, err := NewSQSDeferralService("foo", mockSQS, make(chan Deferral))
		if err != nil {
			t.Fatalf("Test %d: NewSQSDeferralService() unexpected error: %d\n", testNumber, err)
		}

		if err := deferralService.Defer(test.wantProjectKey, test.wantBranch); err != nil {
			t.Fatalf("Test %d: Defer() unexpected error: %d\n", testNumber, err)
		}

		if x, ok := deferralService.(*SQSDeferralService); ok {
			if x.queueURL != mockSQS.queueURL {
				t.Errorf("Want %s, got %s\n", mockSQS.queueURL, x.queueURL)
			}
		} else {
			t.Fatalf("This test assumes a concrete instance of SQSDeferralService, but found %T\n", deferralService)
		}

		if *mockSQS.createQueueInput.QueueName != test.wantQueueName {
			t.Errorf("Test %d: want queueName %s, got %s\n", testNumber, test.wantQueueName, *mockSQS.createQueueInput.QueueName)
		}

		if *mockSQS.messageInput.MessageAttributes["projectkey"].StringValue != test.wantProjectKey {
			t.Errorf("Test %d: want projectkey %s, got %s\n", testNumber, test.wantProjectKey, *mockSQS.messageInput.MessageAttributes["projectKey"].StringValue)
		}

		if *mockSQS.messageInput.MessageAttributes["branch"].StringValue != test.wantBranch {
			t.Errorf("Test %d: want branch %s, got %s\n", testNumber, test.wantBranch, *mockSQS.messageInput.MessageAttributes["branch"].StringValue)
		}

		if *mockSQS.messageInput.MessageAttributes["unixtime"].StringValue == "0" {
			t.Errorf("Test %d: want unix time not zero, got %s\n", testNumber, *mockSQS.messageInput.MessageAttributes["unixtime"].StringValue)
		}
	}
}
