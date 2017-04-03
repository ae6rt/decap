package main

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/retry"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

// AWSStorageService is a container for holding AWS configuration
type AWSStorageService struct {
	Config *aws.Config
}

// NewAWSStorageService returns a StorageService implemented on top of AWS
func NewAWSStorageService(awsCredential AWSCredential) StorageService {
	return AWSStorageService{aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(awsCredential.accessKey, awsCredential.accessSecret, "")).WithRegion(awsCredential.region).WithMaxRetries(3)}
}

// GetBuildsByProject returns logical builds by team / project.
func (c AWSStorageService) GetBuildsByProject(project v1.Project, since uint64, limit uint64) ([]v1.Build, error) {

	var resp *dynamodb.QueryOutput

	work := func() error {
		svc := dynamodb.New(session.New(), c.Config)
		params := &dynamodb.QueryInput{
			TableName:              aws.String("decap-build-metadata"),
			IndexName:              aws.String("project-key-build-start-time-index"),
			KeyConditionExpression: aws.String("#pkey = :pkey and #bst > :since"),
			ExpressionAttributeNames: map[string]*string{
				"#pkey": aws.String("project-key"),
				"#bst":  aws.String("build-start-time"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":pkey": {
					S: aws.String(projectKey(project.Team, project.ProjectName)),
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
	err := retry.New(3, retry.DefaultBackoffFunc).Try(work)
	if err != nil {
		return nil, err
	}

	var builds []v1.Build
	for _, v := range resp.Items {
		buildDuration, err := strconv.ParseUint(*v["build-duration"].N, 10, 64)
		if err != nil {
			Log.Printf("Error converting build-duration to ordinal value: %v\n", err)
		}
		buildResult, err := strconv.ParseInt(*v["build-result"].N, 10, 32)
		if err != nil {
			Log.Printf("Error converting build-result to ordinal value: %v\n", err)
		}
		buildTime, err := strconv.ParseUint(*v["build-start-time"].N, 10, 64)
		if err != nil {
			Log.Printf("Error converting build-start-time to ordinal value: %v\n", err)
		}

		build := v1.Build{
			ID:         *v["build-id"].S,
			ProjectKey: *v["project-key"].S,
			Branch:     *v["branch"].S,
			Duration:   buildDuration,
			Result:     int(buildResult),
			UnixTime:   buildTime,
		}
		builds = append(builds, build)
	}
	return builds, nil
}

// GetArtifacts returns the file manifest of artifacts tar file if the Accept: text/plain header
// is set.  Otherwise returns the build artifacts as a gzipped tar file.
func (c AWSStorageService) GetArtifacts(buildID string) ([]byte, error) {
	return c.bytesFromBucket("decap-build-artifacts", buildID)
}

// GetConsoleLog returns console logs in plain text if the Accept: text/plain header
// is set.  Otherwise returns the console log as a gzipped archive.
func (c AWSStorageService) GetConsoleLog(buildID string) ([]byte, error) {
	return c.bytesFromBucket("decap-console-logs", buildID)
}

func (c AWSStorageService) bytesFromBucket(bucketName, objectKey string) ([]byte, error) {

	var resp *s3.GetObjectOutput

	work := func() error {
		svc := s3.New(session.New(), c.Config)

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

	if err := retry.New(3, retry.DefaultBackoffFunc).Try(work); err != nil {
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}
