package deferrals

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type ResubmitMock struct {
	MockSQS
	queueURL            string
	projectKey          string
	branch              string
	receiptHandle       string
	deleteMessageInput  *sqs.DeleteMessageInput
	createQueueInput    *sqs.CreateQueueInput
	receiveMessageInput *sqs.ReceiveMessageInput
	wg                  *sync.WaitGroup
}

func (t *ResubmitMock) CreateQueue(f *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	t.createQueueInput = f
	return &sqs.CreateQueueOutput{QueueUrl: &t.queueURL}, nil
}

func (t *ResubmitMock) DeleteMessage(f *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	t.deleteMessageInput = f
	defer t.wg.Done()
	return &sqs.DeleteMessageOutput{}, nil
}

func (t *ResubmitMock) ReceiveMessage(f *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	t.receiveMessageInput = f
	return &sqs.ReceiveMessageOutput{
		Messages: []*sqs.Message{
			&sqs.Message{
				MessageAttributes: map[string]*sqs.MessageAttributeValue{
					"projectkey": &sqs.MessageAttributeValue{
						DataType:    aws.String("String"),
						StringValue: aws.String(t.projectKey),
					},
					"branch": &sqs.MessageAttributeValue{
						DataType:    aws.String("String"),
						StringValue: aws.String(t.branch),
					},
					"unixtime": &sqs.MessageAttributeValue{
						DataType:    aws.String("String"),
						StringValue: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
					},
				},
				ReceiptHandle: aws.String(t.receiptHandle),
			},
		},
	}, nil
}

func TestResubmit(t *testing.T) {
	var tests = []struct {
		queueName     string
		queueURL      string
		projectKey    string
		branch        string
		receiptHandle string
		unixtime      int64
	}{
		{
			queueName:     "foo",
			queueURL:      "q",
			projectKey:    "p1",
			branch:        "b1",
			receiptHandle: "r1",
			unixtime:      22,
		},
	}

	for testNumber, test := range tests {
		deferralChannel := make(chan Deferral)

		// Resubmit() deletes the read messages in a goroutine.  Wait for this goroutine to run before testing its expected effects.
		var wg sync.WaitGroup
		wg.Add(1)

		mockSQS := &ResubmitMock{
			projectKey:    test.projectKey,
			branch:        test.branch,
			queueURL:      test.queueURL,
			receiptHandle: test.receiptHandle,
			wg:            &wg,
		}

		deferralService, _ := NewSQSDeferralService(test.queueName, mockSQS, deferralChannel)

		go deferralService.Resubmit()

		deferral := <-deferralChannel

		// wait for delete-message to finish
		wg.Wait()

		// test create service effects
		if x, ok := deferralService.(*SQSDeferralService); ok {
			if x.queueURL != mockSQS.queueURL {
				t.Errorf("Want %s, got %s\n", mockSQS.queueURL, x.queueURL)
			}
		} else {
			t.Fatalf("This test assumes a concrete instance of SQSDeferralService, but found %T\n", deferralService)
		}

		// test recieve message effects
		if deferral.ProjectKey != test.projectKey {
			t.Errorf("Test %d: want %s, got %s\n", testNumber, test.projectKey, deferral.ProjectKey)
		}
		if deferral.Branch != test.branch {
			t.Errorf("Test %d: want %s, got %s\n", testNumber, test.branch, deferral.Branch)
		}
		if deferral.UnixTime == 0 {
			t.Errorf("Test %d: want unixtime > 0, got %d\n", testNumber, test.unixtime)
		}

		// test delete message effects
		if *mockSQS.deleteMessageInput.QueueUrl != test.queueURL {
			t.Errorf("Test %d: want queue URL %s, got %s\n", testNumber, test.queueURL, mockSQS.deleteMessageInput.QueueUrl)
		}

		if *mockSQS.deleteMessageInput.ReceiptHandle != test.receiptHandle {
			t.Errorf("Test %d: want receipt handle %s, got %s\n", testNumber, test.receiptHandle, mockSQS.deleteMessageInput.ReceiptHandle)
		}

		close(deferralChannel)
	}
}
