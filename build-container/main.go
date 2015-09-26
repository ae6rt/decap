package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ae6rt/decap/build-container/locks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

var buildInfo string

var buildStartTime int64
var buildDuration int64
var buildResult int64

var bucketName string
var buildID string
var contentType string
var fileName string
var awsRegion string
var Log *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

var BCToolCmd = &cobra.Command{
	Use:   "bctool",
	Short: "bctool is a multifunction build container tool.",
	Long:  `A multifunction build container tool that unlocks builds, uploads files to S3, and puts items to DynamoDb`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of bctool",
	Long:  `All software has versions. This is bctool's`,
	Run: func(cmd *cobra.Command, args []string) {
		Log.Println(buildInfo)
		os.Exit(0)
	},
}

var unlockBuildCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock a build",
	Long:  `Unlock a build`,
	Run: func(cmd *cobra.Command, args []string) {
		locks.Unlock()
	},
}

var putS3Cmd = &cobra.Command{
	Use:   "s3put",
	Short: "put a file to an S3 bucket",
	Long:  `put a file to an S3 bucket`,
	Run: func(cmd *cobra.Command, args []string) {
		config := aws.NewConfig().WithCredentials(credentials.NewEnvCredentials()).WithRegion(awsRegion).WithMaxRetries(3)
		svc := s3.New(config)
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			Log.Fatal(err)
		}
		params := &s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(buildID),
			Body:          bytes.NewReader(data),
			ContentType:   aws.String(contentType),
			ContentLength: aws.Int64(int64(len(data))),
		}
		if _, err := svc.PutObject(params); err != nil {
			Log.Fatal(err.Error())
		} else {
			Log.Println("S3 Put successful")
		}
	},
}

var buildStartCmd = &cobra.Command{
	Use:   "build-start",
	Short: "Mark a build as started in DynamoDb",
	Long:  "Mark a build as started in DynamoDb.  This sets the isBuilding flag and sets the build start time.",
	Run: func(cmd *cobra.Command, args []string) {
		config := aws.NewConfig().WithCredentials(credentials.NewEnvCredentials()).WithRegion(awsRegion).WithMaxRetries(3)
		svc := dynamodb.New(config)
		params := &dynamodb.PutItemInput{
			TableName: aws.String("decap-build-metadata"),
			Item: map[string]*dynamodb.AttributeValue{
				"buildID": {
					S: aws.String(os.Getenv("BUILD_ID")),
				},
				"projectKey": {
					S: aws.String(os.Getenv("PROJECT_KEY")),
				},
				"buildTime": {
					N: aws.String(fmt.Sprintf("%d", buildStartTime)),
				},
				"branch": {
					S: aws.String(os.Getenv("BRANCH_TO_BUILD")),
				},
				"isBuilding": {
					N: aws.String("1"),
				},
			},
		}
		if _, err := svc.PutItem(params); err != nil {
			Log.Fatal(err.Error())
		} else {
			Log.Println("DynamoDb Put successful")
		}
	},
}

var buildFinishCmd = &cobra.Command{
	Use:   "build-finish",
	Short: "Mark a build as finished in DynamoDb",
	Long:  "Mark a build as finished in DynamoDb.  This clears the isBuilding flag and sets the build result and duration.",
	Run: func(cmd *cobra.Command, args []string) {
		config := aws.NewConfig().WithCredentials(credentials.NewEnvCredentials()).WithRegion(awsRegion).WithMaxRetries(3)
		svc := dynamodb.New(config)
		params := &dynamodb.UpdateItemInput{
			TableName: aws.String("decap-build-metadata"),
			Key: map[string]*dynamodb.AttributeValue{
				"buildID": {
					S: aws.String(os.Getenv("BUILD_ID")),
				},
			},
			UpdateExpression: aws.String("SET buildElapsedTime = :buildDuration, buildResult = :buildResult, isBuilding = :isBuilding"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":buildDuration": {
					N: aws.String(fmt.Sprintf("%d", buildDuration)),
				},
				":buildResult": {
					N: aws.String(fmt.Sprintf("%d", buildResult)),
				},
				":isBuilding": {
					N: aws.String("0"),
				},
			},
		}
		if _, err := svc.UpdateItem(params); err != nil {
			Log.Fatal(err.Error())
		} else {
			Log.Println("DynamoDb Update successful")
		}
	},
}

func init() {
	putS3Cmd.Flags().StringVarP(&bucketName, "bucket-name", "b", "", "S3 Bucket Name")
	putS3Cmd.Flags().StringVarP(&buildID, "build-id", "i", "", "Build ID")
	putS3Cmd.Flags().StringVarP(&contentType, "content-type", "t", "", "Content Type")
	putS3Cmd.Flags().StringVarP(&fileName, "filename", "f", "", "File Name")
	putS3Cmd.Flags().StringVarP(&awsRegion, "aws-region", "r", "us-west-1", "AWS Region")

	buildStartCmd.Flags().StringVarP(&awsRegion, "aws-region", "r", "us-west-1", "AWS Region")
	buildStartCmd.Flags().Int64VarP(&buildStartTime, "start-time", "s", 0, "Unix time in seconds since the epoch when the build started")

	buildFinishCmd.Flags().StringVarP(&awsRegion, "aws-region", "r", "us-west-1", "AWS Region")
	buildFinishCmd.Flags().Int64VarP(&buildResult, "build-result", "s", 0, "Unix exit code of the executed build")
	buildFinishCmd.Flags().Int64VarP(&buildDuration, "build-duration", "d", 0, "Duration of the build in seconds")

	BCToolCmd.AddCommand(versionCmd)
	BCToolCmd.AddCommand(unlockBuildCmd)
	BCToolCmd.AddCommand(putS3Cmd)
	BCToolCmd.AddCommand(buildStartCmd)
	BCToolCmd.AddCommand(buildFinishCmd)
}

func main() {
	BCToolCmd.Execute()
}
