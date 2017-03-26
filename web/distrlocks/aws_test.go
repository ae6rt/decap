package distrlocks

import (
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// A minimal DynamoDB mock that implements as much of the public API we need.
type happyDB struct {
}

func (t *happyDB) PutItem(i *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, nil
}

func (t *happyDB) DeleteItem(i *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return nil, nil
}

type acquireMock struct {
	happyDB
	f *dynamodb.PutItemInput
}

func (t *acquireMock) PutItem(i *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	t.f = i
	return nil, nil
}

type releaseMock struct {
	happyDB
	f *dynamodb.DeleteItemInput
}

func (t *releaseMock) DeleteItem(i *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	t.f = i
	return nil, nil
}

func TestCreateLock(t *testing.T) {
	dbMock := &acquireMock{}
	lockService := NewDynamoDbLockService(dbMock)
	event := v1.UserBuildEvent{Team_: "proj", Project_: "code", Ref_: "feature/foo"}

	if err := lockService.Acquire(event); err != nil {
		t.Errorf("%v\n", err)
	}

	if *dbMock.f.TableName != "decap-buildlocks" {
		t.Errorf("Want decap-buildlocks, got %s\n", *dbMock.f.TableName)
	}

	if *dbMock.f.Item["lockname"].S != "proj/code/feature/foo" {
		t.Errorf("Want proj/feature/foo, got %s\n", *dbMock.f.Item["lockname"].S)
	}
}

func TestReleaseLock(t *testing.T) {
	dbMock := &releaseMock{}
	lockService := NewDynamoDbLockService(dbMock)
	event := v1.UserBuildEvent{Team_: "proj", Project_: "code", Ref_: "feature/foo"}

	if err := lockService.Release(event); err != nil {
		t.Errorf("%v\n", err)
	}

	if *dbMock.f.TableName != "decap-buildlocks" {
		t.Errorf("Want decap-buildlocks, got %s\n", *dbMock.f.TableName)
	}

	if *dbMock.f.Key["lockname"].S != "proj/code/feature/foo" {
		t.Errorf("Want proj/feature/foo, got %s\n", *dbMock.f.Key["lockname"].S)
	}
}
