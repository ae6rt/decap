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
		sqsService := &CreateNewSQSMock{queueURL: test.queueURL}

		s, _ := NewSQSDeferralService(test.queueName, sqsService, nil)

		if x, ok := s.(*SQSDeferralService); ok {
			if x.queueURL != sqsService.queueURL {
				t.Errorf("Want %s, got %s\n", sqsService.queueURL, x.queueURL)
			}
		} else {
			t.Fatalf("This test assumes a concrete instance of SQSDeferralService, but found %T\n", s)
		}

		if *sqsService.param.QueueName != test.queueName {
			t.Errorf("Test %d: want queueName %s, got %s\n", testNumber, test.queueName, sqsService.param.QueueName)
		}
	}
}
