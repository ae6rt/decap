package lock

import (
	"fmt"
	"time"

	"github.com/ae6rt/decap/web/api/v1"
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
func NewDynamoDbLockService(db DynamoDB) LockService {
	return DynamoDbLockService{db: db}
}

// Acquire conditionally puts a lock representation in DynamoDb for the given input lock.
// The operation will fail if the lock exists and is not expired.  A locked is deemed good (unexpired)
// for a finite amount of time, should a process for some reason fail to remove it manually after a
// branch build is completed.
func (l DynamoDbLockService) Acquire(lock v1.UserBuildEvent) error {
	params := &dynamodb.PutItemInput{
		TableName: aws.String("decap-buildlocks"),
		Item: map[string]*dynamodb.AttributeValue{
			"lockname": {
				S: aws.String(lock.Lockname()),
			},
			"expiresunixtime": {
				N: aws.String(fmt.Sprintf("%d", time.Now().Add(3*time.Hour).Unix())),
			},
		},
		ConditionExpression: aws.String("attribute_not_exists(#lockname) OR (#expiresunixtime < :nowunixtime)"),
		ExpressionAttributeNames: map[string]*string{
			"#lockname":        aws.String("lockname"),
			"#expiresunixtime": aws.String("expiresunixtime"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":nowunixtime": {
				N: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
			},
		},
	}

	if _, err := l.db.PutItem(params); err != nil {
		return errors.Wrap(err, fmt.Sprintf("%T: Failed to acquire lock %s", err, lock.Lockname()))
	}

	return nil
}

// Release removes a lock from the lock table.
func (l DynamoDbLockService) Release(lock v1.UserBuildEvent) error {
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String("decap-buildlocks"),
		Key: map[string]*dynamodb.AttributeValue{
			"lockname": {
				S: aws.String(lock.Lockname()),
			},
		},
	}

	if _, err := l.db.DeleteItem(params); err != nil {
		return errors.Wrap(err, "Failed to release lock "+lock.Lockname())
	}

	return nil
}
