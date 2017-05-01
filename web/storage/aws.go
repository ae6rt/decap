package storage

import (
	"fmt"
	"io/ioutil"

	"strconv"

	"github.com/ae6rt/decap/web/api/v1"
	decapcreds "github.com/ae6rt/decap/web/credentials"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

// S3 models the subset of all Amazon S3 that we need.
type S3 interface {
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

// DB models the subset of DynamoDb that we need.
type DB interface {
	Query(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
}

// DefaultStorageService is the default working storage service, which is based on Amazon S3 and DynamoDb.
type DefaultStorageService struct {
	db      DB
	buckets S3
	logger  log.Logger
}

// NewAWS returns a StorageService implemented on top of Amazon S3 and DynamoDb.
func NewAWS(credential decapcreds.AWSCredential, logger log.Logger) Service {
	sess := session.Must(session.NewSession(
		aws.NewConfig().WithCredentials(
			credentials.NewStaticCredentials(credential.AccessKey, credential.AccessSecret, ""),
		).WithRegion(credential.Region).WithMaxRetries(3)),
	)

	return DefaultStorageService{
		db:      dynamodb.New(sess),
		buckets: s3.New(sess),
		logger:  logger,
	}
}

// GetBuildsByProject returns logical builds by team / project.
func (c DefaultStorageService) GetBuildsByProject(project v1.Project, since uint64, limit uint64) ([]v1.Build, error) {
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
				S: aws.String(project.Key()),
			},
			":since": {
				N: aws.String(fmt.Sprintf("%d", since)),
			},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int64(int64(limit)),
	}

	resp, err := c.db.Query(params)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error retrieving build information for project %s\n", project.Key()))
	}

	var builds []v1.Build
	for _, v := range resp.Items {
		buildDuration, err := strconv.ParseUint(*v["build-duration"].N, 10, 64)
		if err != nil {
			c.logger.Log("Error converting build-duration to ordinal value: %v\n", err)
		}
		buildResult, err := strconv.ParseInt(*v["build-result"].N, 10, 32)
		if err != nil {
			c.logger.Log("Error converting build-result to ordinal value: %v\n", err)
		}
		buildTime, err := strconv.ParseUint(*v["build-start-time"].N, 10, 64)
		if err != nil {
			c.logger.Log("Error converting build-start-time to ordinal value: %v\n", err)
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
func (c DefaultStorageService) GetArtifacts(buildID string) ([]byte, error) {
	return c.bytesFromBucket("decap-build-artifacts", buildID)
}

// GetConsoleLog returns console logs in plain text if the Accept: text/plain header
// is set.  Otherwise returns the console log as a gzipped archive.
func (c DefaultStorageService) GetConsoleLog(buildID string) ([]byte, error) {
	return c.bytesFromBucket("decap-console-logs", buildID)
}

func (c DefaultStorageService) bytesFromBucket(bucketName, objectKey string) ([]byte, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	resp, err := c.buckets.GetObject(params)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error retrieving object %s from bucket %s\n", bucketName, objectKey))
	}

	return ioutil.ReadAll(resp.Body)
}
