package deferrals

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
)

type DeferMock struct {
	MockSQS
	queueName        string
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
		sqsService := &DeferMock{queueURL: test.wantQueueURL}

		s, err := NewSQSDeferralService("foo", sqsService, make(chan Deferral))
		if err != nil {
			t.Fatalf("Test %d: NewSQSDeferralService() unexpected error: %d\n", testNumber, err)
		}

		if err := s.Defer(test.wantProjectKey, test.wantBranch); err != nil {
			t.Fatalf("Test %d: Defer() unexpected error: %d\n", testNumber, err)
		}

		if x, ok := s.(*SQSDeferralService); ok {
			if x.queueURL != sqsService.queueURL {
				t.Errorf("Want %s, got %s\n", sqsService.queueURL, x.queueURL)
			}
		} else {
			t.Fatalf("This test assumes a concrete instance of SQSDeferralService, but found %T\n", s)
		}

		if *sqsService.createQueueInput.QueueName != test.wantQueueName {
			t.Errorf("Test %d: want queueName %s, got %s\n", testNumber, test.wantQueueName, *sqsService.createQueueInput.QueueName)
		}

		if *sqsService.messageInput.MessageAttributes["projectkey"].StringValue != test.wantProjectKey {
			t.Errorf("Test %d: want projectkey %s, got %s\n", testNumber, test.wantProjectKey, *sqsService.messageInput.MessageAttributes["projectKey"].StringValue)
		}

		if *sqsService.messageInput.MessageAttributes["branch"].StringValue != test.wantBranch {
			t.Errorf("Test %d: want branch %s, got %s\n", testNumber, test.wantBranch, *sqsService.messageInput.MessageAttributes["branch"].StringValue)
		}

		if *sqsService.messageInput.MessageAttributes["unixtime"].StringValue == "0" {
			t.Errorf("Test %d: want unix time not zero, got %s\n", testNumber, *sqsService.messageInput.MessageAttributes["unixtime"].StringValue)
		}
	}
}
