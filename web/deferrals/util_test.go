package deferrals

import (
	"os"

	"github.com/aws/aws-sdk-go/service/sqs"
)

type MockSQS struct {
}

func (t MockSQS) CreateQueue(f *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	return nil, nil
}

func (t MockSQS) DeleteMessage(f *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return nil, nil
}

func (t MockSQS) ReceiveMessage(f *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	return nil, nil
}

func (t MockSQS) SendMessage(f *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return nil, nil
}

func awsCoordinates() (string, string, string) {
	key := os.Getenv("AWS_ACCESS_KEY_ID")
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_DEFAULT_REGION")

	if key == "" || secret == "" || region == "" {
		panic("AWS key, secret, and region expected in the environment at AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_DEFAULT_REGION, respectively.")
	}

	return key, secret, region
}
