package deferrals

import (
	"fmt"
	"log"
	"sync"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

// DynamoDB models the AWS dynamodb service.
type DynamoDB interface {
	PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
}

// DynamoDBDeferralService is the working network deferral service.
type DynamoDBDeferralService struct {
	deferralTable string
	mutex         sync.Mutex
	db            DynamoDB
	relay         chan<- v1.UserBuildEvent
	logger        *log.Logger
}

// NewDynamoDBDeferralService is the constructor for a DeferralService built on top of DynamoDB.
func NewDynamoDBDeferralService(deferralTable string, db DynamoDB, r chan<- v1.UserBuildEvent, log *log.Logger) DeferralService {
	return &DynamoDBDeferralService{deferralTable: deferralTable, db: db, relay: r, logger: log}
}

// NewDynamoDB creates a new network client for interacting with Amazon DynamoDB
func NewDynamoDB(awsAccessKey, awsAccessSecret, awsRegion string) DynamoDB {
	sess := session.Must(session.NewSession(aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(awsAccessKey, awsAccessSecret, "")).WithRegion(awsRegion).WithMaxRetries(3)))
	return dynamodb.New(sess)
}

// Defer defers a build.
func (t *DynamoDBDeferralService) Defer(buildEvent v1.UserBuildEvent) error {
	params := &dynamodb.PutItemInput{
		TableName: aws.String("decap-deferrals"),
		Item: map[string]*dynamodb.AttributeValue{
			"project-key": {
				S: aws.String(buildEvent.Key()),
			},
			"build-id": {
				S: aws.String(buildEvent.ID),
			},
			"deferred-unixtime": {
				N: aws.String(fmt.Sprintf("%d", buildEvent.DeferredUnixtime)),
			},
		},
	}

	// TODO do we really need to acquire the mutex here?  Probably.
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, err := t.db.PutItem(params); err != nil {
		return errors.Wrap(err, "error deferring build "+buildEvent.ID)
	}

	t.logger.Printf("Build %s deferred", buildEvent.ID)
	return nil
}

// List returns the list of current deferrals.
func (t *DynamoDBDeferralService) List() ([]v1.UserBuildEvent, error) {
	return nil, nil
}

// Removes all deferred builds where project + branch matches key.
func (t *DynamoDBDeferralService) Remove(key string) error {
	return nil
}

// Resubmit reads deferred builds and resubmits them for launching.
func (t *DynamoDBDeferralService) Resubmit() {
}
