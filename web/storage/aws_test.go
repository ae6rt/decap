package storage

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

type MockDynamoDB struct {
	captureInput *dynamodb.QueryInput
	output       *dynamodb.QueryOutput
	forceError   bool
}

func (t *MockDynamoDB) Query(in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	t.captureInput = in
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return t.output, err
}

type MockS3 struct {
	captureInput *s3.GetObjectInput
	output       *s3.GetObjectOutput
	forceError   bool
}

func (t *MockS3) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	t.captureInput = in
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return t.output, err
}

func TestDynamoDbGetBuilds(t *testing.T) {
	var tests = []struct {
		inProject  v1.Project
		inSince    uint64
		inLimit    int64
		output     *dynamodb.QueryOutput
		want       []v1.Build
		forceError bool
	}{
		{
			inProject: v1.Project{
				Team:        "ae6rt",
				ProjectName: "proj",
			},
			inSince: 0,
			inLimit: 100,
			output: &dynamodb.QueryOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					map[string]*dynamodb.AttributeValue{
						"build-id":         &dynamodb.AttributeValue{S: aws.String("id")},
						"project-key":      &dynamodb.AttributeValue{S: aws.String("ae6rt/proj")},
						"branch":           &dynamodb.AttributeValue{S: aws.String("issue/32")},
						"build-duration":   &dynamodb.AttributeValue{N: aws.String("2")},
						"build-result":     &dynamodb.AttributeValue{N: aws.String("0")},
						"build-start-time": &dynamodb.AttributeValue{N: aws.String("99")},
					},
				},
			},
			want: []v1.Build{
				v1.Build{
					ID:         "id",
					ProjectKey: "ae6rt/proj",
					Branch:     "issue/32",
					Duration:   2,
					Result:     0,
					UnixTime:   99,
				},
			},
		},
		{
			forceError: true,
		},
	}

	for testNumber, test := range tests {
		db := &MockDynamoDB{output: test.output, forceError: test.forceError}
		service := DefaultStorageService{db: db}
		got, err := service.GetBuildsByProject(test.inProject, test.inSince, uint64(test.inLimit))
		if test.forceError {
			if err == nil {
				t.Errorf("Test %d: expecting an error\n", testNumber)
			}
			continue
		} else if err != nil {
			t.Errorf("Test %d: unexpected error %v\n", testNumber, err)
		}

		if len(got) != len(test.want) {
			t.Errorf("Test %d: GetBuildsByProject(%v,%d,%d) want %+v, got %+v\n", testNumber, test.inProject, test.inSince, test.inLimit, test.want, got)
		}

		for k, v := range test.want {
			if v.ID != got[k].ID {
				t.Errorf("Test %d: want %v, got %v\n", testNumber, v.ID, got[k].ID)
			}
			if v.ProjectKey != got[k].ProjectKey {
				t.Errorf("Test %d: want %v, got %v\n", testNumber, v.ProjectKey, got[k].ProjectKey)
			}
			if v.Branch != got[k].Branch {
				t.Errorf("Test %d: want %v, got %v\n", testNumber, v.Branch, got[k].Branch)
			}
			if v.Duration != got[k].Duration {
				t.Errorf("Test %d: want %v, got %v\n", testNumber, v.Duration, got[k].Duration)
			}
			if v.Result != got[k].Result {
				t.Errorf("Test %d: want %v, got %v\n", testNumber, v.Result, got[k].Result)
			}
			if v.UnixTime != got[k].UnixTime {
				t.Errorf("Test %d: want %v, got %v\n", testNumber, v.UnixTime, got[k].UnixTime)
			}
		}

		// Verify input capture
		if *db.captureInput.ExpressionAttributeNames["#pkey"] != "project-key" {
			t.Errorf("Test %d: want #pkey = %s, got %s\n", testNumber, "project-key", *db.captureInput.ExpressionAttributeNames["#pkey"])
		}
		if *db.captureInput.ExpressionAttributeNames["#bst"] != "build-start-time" {
			t.Errorf("Test %d: want #bst = %s, got %s\n", testNumber, "build-start-time", *db.captureInput.ExpressionAttributeNames["#bst"])
		}

		if *db.captureInput.ExpressionAttributeValues[":pkey"].S != test.inProject.Key() {
			t.Errorf("Test %d: want :pkey = %s, got %s\n", testNumber, test.inProject.Key(), *db.captureInput.ExpressionAttributeValues[":pkey"].S)
		}
		if *db.captureInput.ExpressionAttributeValues[":since"].N != fmt.Sprintf("%d", test.inSince) {
			t.Errorf("Test %d: want :since = %s, got %s\n", testNumber, fmt.Sprintf("%d", test.inSince), *db.captureInput.ExpressionAttributeValues[":since"].N)
		}

		if *db.captureInput.Limit != test.inLimit {
			t.Errorf("Test %d: want Limit = %d, got %d\n", testNumber, test.inLimit, *db.captureInput.Limit)
		}
	}
}

func TestS3(t *testing.T) {
	var tests = []struct {
		inBuildID  string
		bucketName string
		output     *s3.GetObjectOutput
		want       []byte
		forceError bool
	}{
		{
			bucketName: "decap-build-artifacts",
			inBuildID:  "id",
			output: &s3.GetObjectOutput{
				Body: ioutil.NopCloser(bytes.NewBuffer([]byte{2})),
			},
			want: []byte{2},
		},
		{
			bucketName: "decap-console-logs",
			inBuildID:  "id",
			output: &s3.GetObjectOutput{
				Body: ioutil.NopCloser(bytes.NewBuffer([]byte{2})),
			},
			want: []byte{2},
		},
		{
			bucketName: "decap-console-logs",
			forceError: true,
		},
		{
			bucketName: "decap-build-artifacts",
			forceError: true,
		},
	}

	for testNumber, test := range tests {
		buckets := &MockS3{output: test.output, forceError: test.forceError}
		service := DefaultStorageService{buckets: buckets}

		var data []byte
		var err error

		switch test.bucketName {
		case "decap-build-artifacts":
			data, err = service.GetArtifacts(test.inBuildID)
		case "decap-console-logs":
			data, err = service.GetConsoleLog(test.inBuildID)
		}

		if test.forceError {
			if err == nil {
				t.Errorf("Test %d: expecting an error\n", testNumber)
			}
			continue
		} else if err != nil {
			t.Errorf("Test %d: unexpected error %v\n", testNumber, err)
			continue
		}

		if len(data) != len(test.want) {
			t.Errorf("Test %d, want %d, got %d\n", testNumber, len(test.want), len(data))
		}

		for k, v := range test.want {
			if data[k] != v {
				t.Errorf("Test %d: want test.want[%d]=%v, got %v\n", testNumber, k, v, data[k])
			}
		}

		// Verify capture input
		if *buckets.captureInput.Bucket != test.bucketName {
			t.Errorf("Test %d: want bucket name == %s, got %s\n", testNumber, test.bucketName, *buckets.captureInput.Bucket)
		}
		if *buckets.captureInput.Key != test.inBuildID {
			t.Errorf("Test %d: want bucket key %s, got %s\n", testNumber, test.inBuildID, *buckets.captureInput.Key)
		}
	}
}
