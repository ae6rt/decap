package distrlocks

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

// DynamoDB is a minimal DynamoDB interface
type DynamoDB interface {
	PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	DeleteItem(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
}

// DynamoDbLockService implements a distributed lock service on top of Amazon DynamoDB.
type DynamoDbLockService struct {
	db DynamoDB
}

// NewDynamoDB returns a new DynamoDB instance initialized with the give access key, secret,and region.
func NewDynamoDB(awsAccessKey, awsAccessSecret, awsRegion string) DynamoDB {
	sess := session.Must(session.NewSession(aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(awsAccessKey, awsAccessSecret, "")).WithRegion(awsRegion).WithMaxRetries(3)))
	return dynamodb.New(sess)
}

// NewDynamoDbLockService creates a new distributed lock service on top of DynamoDb.
func NewDynamoDbLockService(db DynamoDB) DistributedLockService {
	return DynamoDbLockService{db: db}
}

// Acquire conditionally puts a lock representation DynamoDb for the given input lock.
// The operation will fail if the lock exists and is not expired.
func (l DynamoDbLockService) Acquire(lock DistributedLock) error {
	params := &dynamodb.PutItemInput{
		TableName: aws.String("decap-buildlocks"),
		Item: map[string]*dynamodb.AttributeValue{
			"lockname": {
				S: aws.String(lock.Key()),
			},
			"expiresunixtime": {
				N: aws.String(fmt.Sprintf("%d", lock.Expires)),
			},
		},
		ConditionExpression: aws.String("attribute_not_exists(#name) OR (#expiresunixtime < :nowunixtime)"),
		ExpressionAttributeNames: map[string]*string{
			"#name":            aws.String("lockname"),
			"#expiresunixtime": aws.String("expiresunixtime"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":nowunixtime": {
				N: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
			},
		},
	}

	if _, err := l.db.PutItem(params); err != nil {
		return errors.Wrap(err, "Failed to acquire lock "+lock.Key())
	}
	return nil
}

// Release removes a lock from the lock table.
func (l DynamoDbLockService) Release(lock DistributedLock) error {
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String("decap-buildlocks"),
		Key: map[string]*dynamodb.AttributeValue{
			"lockname": {
				S: aws.String(lock.Key()),
			},
		},
	}

	if _, err := l.db.DeleteItem(params); err != nil {
		return errors.Wrap(err, "Failed to release lock "+lock.Key())
	}

	return nil
}
