package deferrals

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type CreateNewSQSMock struct {
	MockSQS
	param    *sqs.CreateQueueInput
	queueURL string
}

func (t *CreateNewSQSMock) CreateQueue(f *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	t.param = f
	return &sqs.CreateQueueOutput{QueueUrl: aws.String(t.queueURL)}, nil
}

func TestCreateNewSQS(t *testing.T) {
	var tests = []struct {
		queueName string
		queueURL  string
	}{
		{
			queueName: "foo",
			queueURL:  "http://example.com",
		},
	}

	for testNumber, test := range tests {
		mockSQS := &CreateNewSQSMock{queueURL: test.queueURL}

		deferralService, _ := NewSQSDeferralService(test.queueName, mockSQS, nil)

		if x, ok := deferralService.(*SQSDeferralService); ok {
			if x.queueURL != mockSQS.queueURL {
				t.Errorf("Want %s, got %s\n", mockSQS.queueURL, x.queueURL)
			}
		} else {
			t.Fatalf("This test assumes a concrete instance of SQSDeferralService, but found %T\n", deferralService)
		}

		if *mockSQS.param.QueueName != test.queueName {
			t.Errorf("Test %d: want queueName %s, got %s\n", testNumber, test.queueName, mockSQS.param.QueueName)
		}
	}
}
