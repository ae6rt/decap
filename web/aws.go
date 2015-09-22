package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/ae6rt/retry"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AWSStorageService struct {
	Config *aws.Config
}

func NewAWSStorageService(key, secret, region string) StorageService {
	return AWSStorageService{aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(key, secret, "")).WithRegion(region).WithMaxRetries(3)}
}

func (c AWSStorageService) GetBuildsByProject(project Project, since uint64, limit uint64) ([]Build, error) {

	var resp *dynamodb.QueryOutput

	work := func() error {
		svc := dynamodb.New(c.Config)
		params := &dynamodb.QueryInput{
			TableName:              aws.String("decap-build-metadata"),
			IndexName:              aws.String("projectKey-buildTime-index"),
			KeyConditionExpression: aws.String("projectKey = :pkey and buildTime > :since"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":pkey": {
					S: aws.String(projectKey(project.Team, project.Library)),
				},
				":since": {
					N: aws.String(fmt.Sprintf("%d", since)),
				},
			},
			ScanIndexForward: aws.Bool(false),
			Limit:            aws.Int64(int64(limit)),
		}

		var err error
		resp, err = svc.Query(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				Log.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
				if reqErr, ok := err.(awserr.RequestFailure); ok {
					Log.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
				}
			} else {
				Log.Println(err.Error())
			}
			return err
		}
		return nil
	}
	err := retry.New(5*time.Second, 3, retry.DefaultBackoffFunc).Try(work)
	if err != nil {
		return nil, err
	}

	builds := make([]Build, 0)
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

		build := Build{
			ID:       *v["buildID"].S,
			Branch:   *v["branch"].S,
			Duration: buildElapsedTime,
			Result:   int(buildResult),
			UnixTime: buildTime,
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (c AWSStorageService) GetBuildsBuildling() ([]Build, error) {

	var resp *dynamodb.QueryOutput

	work := func() error {
		svc := dynamodb.New(c.Config)
		params := &dynamodb.QueryInput{
			TableName:              aws.String("decap-build-metadata"),
			IndexName:              aws.String("isBuilding-index"),
			KeyConditionExpression: aws.String("isBuilding = :val"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":val": {
					N: aws.String("1"),
				},
			},
			ScanIndexForward: aws.Bool(false),
		}

		var err error
		resp, err = svc.Query(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				Log.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
				if reqErr, ok := err.(awserr.RequestFailure); ok {
					Log.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
				}
			} else {
				Log.Println(err.Error())
			}
			return err
		}
		return nil
	}
	err := retry.New(5*time.Second, 3, retry.DefaultBackoffFunc).Try(work)
	if err != nil {
		return nil, err
	}

	log.Println(awsutil.Prettify(resp))

	builds := make([]Build, 0)
	for _, v := range resp.Items {
		buildTime, err := strconv.ParseUint(*v["buildTime"].N, 10, 64)
		if err != nil {
			Log.Printf("Error converting buildTime to ordinal value: %v\n", err)
		}

		build := Build{
			ProjectKey: *v["projectKey"].S,
			ID:         *v["buildID"].S,
			Branch:     *v["branch"].S,
			UnixTime:   buildTime,
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (c AWSStorageService) GetArtifacts(buildID string) ([]byte, error) {
	return c.bytesFromBucket("decap-build-artifacts", buildID)
}

func (c AWSStorageService) GetConsoleLog(buildID string) ([]byte, error) {
	return c.bytesFromBucket("decap-console-logs", buildID)
}

func (c AWSStorageService) bytesFromBucket(bucketName, objectKey string) ([]byte, error) {

	var resp *s3.GetObjectOutput

	work := func() error {
		svc := s3.New(c.Config)

		params := &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		}

		var err error
		if resp, err = svc.GetObject(params); err != nil {
			Log.Println(err.Error())
			return err
		}
		return nil
	}

	if err := retry.New(5*time.Second, 3, retry.DefaultBackoffFunc).Try(work); err != nil {
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}
