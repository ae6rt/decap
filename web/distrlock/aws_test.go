package distrlocks

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// A minimal DynamoDB mock that implements as much of the public API we need.
type happyDB struct {
}

func (mock *happyDB) PutItem(i *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, nil
}

func (mock *happyDB) DeleteItem(i *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return nil, nil
}

type acquireMock struct {
	happyDB
	f *dynamodb.PutItemInput
}

func (mock *acquireMock) PutItem(i *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	mock.f = i
	return nil, nil
}

type releaseMock struct {
	happyDB
	f *dynamodb.DeleteItemInput
}

func (mock *releaseMock) DeleteItem(i *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	mock.f = i
	return nil, nil
}

func TestCreateLock(t *testing.T) {
	mock := &acquireMock{}
	s := NewDynamoDbLockService(mock)
	l := NewDistributedLock("proj", "feature/foo")

	if err := s.Acquire(l); err != nil {
		t.Errorf("%v\n", err)
	}

	if *mock.f.TableName != "decap-buildlocks" {
		t.Errorf("Want decap-buildlocks, got %s\n", *mock.f.TableName)
	}

	if *mock.f.Item["lockname"].S != "proj/feature/foo" {
		t.Errorf("Want proj/feature/foo, got %s\n", *mock.f.Item["lockname"].S)
	}
}

func TestReleaseLock(t *testing.T) {
	mock := &releaseMock{}
	s := NewDynamoDbLockService(mock)
	l := NewDistributedLock("proj", "feature/foo")

	if err := s.Release(l); err != nil {
		t.Errorf("%v\n", err)
	}

	if *mock.f.TableName != "decap-buildlocks" {
		t.Errorf("Want decap-buildlocks, got %s\n", *mock.f.TableName)
	}

	if *mock.f.Key["lockname"].S != "proj/feature/foo" {
		t.Errorf("Want proj/feature/foo, got %s\n", *mock.f.Key["lockname"].S)
	}
}
