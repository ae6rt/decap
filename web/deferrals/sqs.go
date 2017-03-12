package deferrals

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/ae6rt/decap/web/uuid"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
)

// SQS models the subset of exported methods we need from the greater Amazon SQS interface.
type SQS interface {
	CreateQueue(*sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error)
	DeleteMessage(*sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	ReceiveMessage(*sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}

// SQSDeferralService implements a deferral service on top of Amazon Simple Queue Service.
type SQSDeferralService struct {
	q        SQS
	queueURL string
	relay    chan<- Deferral
}

// NewSQS creates a new network client for interacting with Amazon SQS.
func NewSQS(awsAccessKey, awsAccessSecret, awsRegion string) SQS {
	sess := session.Must(session.NewSession(aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(awsAccessKey, awsAccessSecret, "")).WithRegion(awsRegion).WithMaxRetries(3)))
	return sqs.New(sess)
}

// NewSQSDeferralService returns a new build deferral service based on Amazon SQS.  The write-only channel of Deferral is
// used to send Deferral events to an actor that relaunches the deferred build.  Those origin of those events
// are deferral messages on the SQS message bus.
func NewSQSDeferralService(s SQS, r chan<- Deferral) DeferralService {
	return &SQSDeferralService{q: s, relay: r}
}

// CreateQueue creates a deferral queue with the name queueName.
func (s *SQSDeferralService) CreateQueue(queueName string) error {
	params := &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	}

	resp, err := s.q.CreateQueue(params)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("%T: Error creating queue: %s", err, queueName))
	}

	s.queueURL = *resp.QueueUrl
	return nil
}

// Resubmit receives messages from the deferral queue and submits them for reexecution.  It is intended
// for this method to be called by a recurring timer.
func (s *SQSDeferralService) Resubmit() {
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: aws.Int64(10),
		MessageAttributeNames: []*string{
			aws.String("All"),
		},
	}
	resp, err := s.q.ReceiveMessage(params)
	if err != nil {
		log.Println(err)
		return
	}

	var msgs []Deferral
	for _, j := range resp.Messages {
		fmt.Printf("@@@ Resubmit() received message: %+v\n", *j.Body)

		t := j.MessageAttributes["unixtime"].StringValue
		unixtime, err := strconv.ParseInt(*t, 10, 64)
		if err != nil {
			log.Printf("Cannot parse unix time in Deferral:  %s\n", t)
			continue
		}

		// might want to sort by unixtime and dedup before sending to channel
		d := Deferral{
			ProjectKey: *j.MessageAttributes["projectkey"].StringValue,
			Branch:     *j.MessageAttributes["branch"].StringValue,
			UnixTime:   unixtime,
		}

		msgs = append(msgs, d)

		// todo: should we bother retrying the delete()? If it fails, the worst that can happen
		// is that a build gets requeued on the next queue read.
		h := j.ReceiptHandle
		go func() {
			if err := s.delete(*h); err != nil {
				log.Printf("Error deleting message %s: %v\n", *h, err)
			}
		}()
	}

	sort.Sort(ByTime(msgs))
	msgs = dedup(msgs)
	for _, d := range msgs {
		s.relay <- d
	}
}

// Defer defers a build based on project key and branch.
func (s *SQSDeferralService) Defer(projectKey, branch string) error {
	params := &sqs.SendMessageInput{
		QueueUrl:     aws.String(s.queueURL),
		MessageBody:  aws.String(uuid.Uuid()),
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

func (s *SQSDeferralService) delete(handle string) error {
	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(handle),
	}

	_, err := s.q.DeleteMessage(params)

	return errors.Wrap(err, "Failed to delete message")
}

// ByTime is used to sort Deferrals by Unix time stamp.
type ByTime []Deferral

func (s ByTime) Len() int {
	return len(s)
}
func (s ByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByTime) Less(i, j int) bool {
	return s[i].UnixTime < s[j].UnixTime
}

func dedup(a []Deferral) []Deferral {
	var results []Deferral

	last := ""
	for _, v := range a {
		if v.Key() != last {
			results = append(results, v)
			last = v.Key()
		}
	}
	return results
}
