package deferrals

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ae6rt/decap/web/uuid"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
)

// SQS models the subset of exported methods we need from the Amazon SQS interface.
type SQS interface {
	CreateQueue(*sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error)
	ReceiveMessage(*sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}

// SQSDeferralService implements a deferral service on top of Amazon Simple Queue Service.
type SQSDeferralService struct {
	q        SQS
	queueURL string
	relay    chan Deferral
}

// createQueue creates a FIFO deferral queue with the name queueName.
func (s *SQSDeferralService) createQueue(queueName string) error {
	params := &sqs.CreateQueueInput{
		QueueName: aws.String("queueName"),
	}

	resp, err := s.q.CreateQueue(params)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("%T: Error creating queue: %s", err, queueName))
	}

	s.queueURL = *resp.QueueUrl
	return nil
}

// Resubmit receives messages from the deferral queue and submits them for reexecution.
func (s *SQSDeferralService) Resubmit() {
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: aws.Int64(10),
	}
	resp, err := s.q.ReceiveMessage(params)
	if err != nil {
		log.Println(err)
		return
	}

	for _, j := range resp.Messages {
		t := *j.MessageAttributes["unixtime"].StringValue
		unixtime, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			log.Printf("Cannot parse unix time in Deferral:  %s\n", t)
			continue
		}

		// might want to sort by unixtime and dedup before sending to channel
		s.relay <- Deferral{
			ProjectKey: *j.MessageAttributes["projectkey"].StringValue,
			Branch:     *j.MessageAttributes["branch"].StringValue,
			UnixTime:   unixtime,
		}
	}
}

// Defer defers a build based on project key and branch.
func (s *SQSDeferralService) Defer(projectKey, branch string) error {
	params := &sqs.SendMessageInput{
		MessageBody:  aws.String(uuid.Uuid()),
		QueueUrl:     aws.String(s.queueURL),
		DelaySeconds: aws.Int64(1),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"projectkey": {
				DataType:    aws.String("String"),
				StringValue: aws.String(projectKey),
			},
			"branch": {
				DataType:    aws.String("String"),
				StringValue: aws.String(branch),
			},
			"unixtime": {
				DataType:    aws.String("Number"),
				StringValue: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
			},
		},
	}
	_, err := s.q.SendMessage(params)
	return err
}
