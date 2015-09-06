package main

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/ae6rt/decap/api/v1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type AWSClient interface {
	GetBuilds(pageStart, pageLimit int) ([]v1.Build, error)
	GetBuildsByProject(project v1.Project, pageStart, pageLimit int) ([]v1.Build, error)
	GetArtifacts(buildID string) ([]byte, error)
	GetConsoleLog(buildID string) ([]byte, error)
}

type DefaultAWSClient struct {
	AccessKeyId string
	SecretKeyId string
	AWSClient
}

func NewDefaultAWSClient() AWSClient {
	key, err := ioutil.ReadFile("/etc/secretes/aws-key")
	if err != nil {
		Log.Printf("No /etc/secrets/aws-key.  Falling back to main default\n", err)
	} else {
		key = []byte(*awsAccessKey)
	}

	secret, err := ioutil.ReadFile("/etc/secretes/aws-secret")
	if err != nil {
		Log.Printf("No /etc/secrets/aws-secret.  Falling back to main default\n", err)
	} else {
		secret = []byte(*awsSecret)
	}

	return DefaultAWSClient{AccessKeyId: string(key), SecretKeyId: string(secret)}
}

func (c DefaultAWSClient) GetBuilds(pageStart, pageLimit int) ([]v1.Build, error) {
	return nil, nil
}

func (c DefaultAWSClient) GetBuildsByProject(project v1.Project, pageStart, pageLimit int) ([]v1.Build, error) {
	config := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(c.AccessKeyId, c.SecretKeyId, "")).WithRegion("us-west-1").WithMaxRetries(3)

	svc := dynamodb.New(config)
	params := &dynamodb.QueryInput{
		TableName:              aws.String("decap-build-metadata"),
		IndexName:              aws.String("projectKey-buildTime-index"),
		KeyConditionExpression: aws.String("projectKey = :pkey"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pkey": {
				S: aws.String(project.Key),
			},
		},
	}

	resp, err := svc.Query(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Generic AWS error with Code, Message, and original error (if any)
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// A service error occurred
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			// This case should never be hit, the SDK should always return an
			// error which satisfies the awserr.Error interface.
			fmt.Println(err.Error())
		}
		return nil, err
	}

	builds := make([]v1.Build, 0)
	for _, v := range resp.Items {
		buildElapsedTime, err := strconv.ParseUint(*v["buildElapsedTime"].N, 10, 64)
		if err != nil {
			Log.Printf("Error converting buildElapsedTime to ordinal value: %v\n", err)
		}
		buildResult, err := strconv.ParseInt(*v["buildResult"].N, 10, 32)
		if err != nil {
			Log.Printf("Error converting buildResult to ordinal value: %v\n", err)
		}
		buildTime, err := strconv.ParseUint(*v["buildTime"].N, 10, 64)
		if err != nil {
			Log.Printf("Error converting buildTime to ordinal value: %v\n", err)
		}

		build := v1.Build{
			ID:       *v["buildID"].S,
			Branch:   *v["branch"].S,
			Duration: buildElapsedTime,
			Result:   buildResult,
			UnixTime: buildTime,
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (c DefaultAWSClient) GetArtifacts(buildID string) ([]byte, error) {
	return nil, nil
}

func (c DefaultAWSClient) GetConsoleLogs(buildID string) ([]byte, error) {
	return nil, nil
}
