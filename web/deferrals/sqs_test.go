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
	sqsService := &CreateNewSQSMock{queueURL: "http://www.example.com"}

	s, _ := NewSQSDeferralService("foo", sqsService, nil)

	if x, ok := s.(*SQSDeferralService); ok && x.queueURL != sqsService.queueURL {
		t.Errorf("Want %s, got %s\n", sqsService.queueURL, x.queueURL)
	}

	if *sqsService.param.QueueName != "foo" {
		t.Errorf("Want foo, got %s\n", sqsService.param.QueueName)
	}

}
