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

var debug bool

var buildID string

var lockServiceBaseURL string
var buildLockKey string

var tableName string
var buildStartTime int64
var buildDuration int64
var buildResult int64
var projectKey string
var branchToBuild string

var bucketName string
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
	Short: "Print version and build info for bctool",
	Long:  `All software has build info. This is bctool's`,
	Run: func(cmd *cobra.Command, args []string) {
		Log.Println(buildInfo)
	},
}

var unlockBuildCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock a build",
	Long:  `Unlock a build`,
	Run: func(cmd *cobra.Command, args []string) {
		locks.Unlock(lockServiceBaseURL, buildID, buildLockKey)
	},
}

var putS3Cmd = &cobra.Command{
	Use:   "s3put",
	Short: "put a file to an S3 bucket",
	Long:  `put a file to an S3 bucket`,
	Run: func(cmd *cobra.Command, args []string) {
		config := aws.NewConfig().WithCredentials(credentials.NewEnvCredentials()).WithRegion(awsRegion).WithMaxRetries(3)
		if debug {
			Log.Printf("%+v\n", config)
		}
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
		if debug {
			Log.Printf("%+v\n", params)
		}
		if resp, err := svc.PutObject(params); err != nil {
			Log.Printf("%+v\n", resp)
			Log.Fatal(err.Error())
		} else {
			if debug {
				Log.Printf("%+v\n", resp)
			}
			Log.Println("S3 Put successful")
		}
	},
}

var recordBuildCmd = &cobra.Command{
	Use:   "record-build-metadata",
	Short: "Record build metadata in backing store",
	Long:  "Record build metadata in backing store",
	Run: func(cmd *cobra.Command, args []string) {
		config := aws.NewConfig().WithCredentials(credentials.NewEnvCredentials()).WithRegion(awsRegion).WithMaxRetries(3)
		if debug {
			Log.Printf("%+v\n", config)
		}
		svc := dynamodb.New(config)
		params := &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"build-id": {
					S: aws.String(buildID),
				},
				"project-key": {
					S: aws.String(projectKey),
				},
				"build-start-time": {
					N: aws.String(fmt.Sprintf("%d", buildStartTime)),
				},
				"branch": {
					S: aws.String(branchToBuild),
				},
				"build-result": {
					N: aws.String(fmt.Sprintf("%d", buildResult)),
				},
				"build-duration": {
					N: aws.String(fmt.Sprintf("%d", buildDuration)),
				},
			},
		}
		if debug {
			Log.Printf("%+v\n", params)
		}
		if resp, err := svc.PutItem(params); err != nil {
			Log.Printf("%+v\n", resp)
			Log.Fatal(err.Error())
		} else {
			if debug {
				Log.Printf("%+v\n", resp)
			}
			Log.Println("DynamoDb Put successful")
		}
	},
}

func main() {
	unlockBuildCmd.Flags().StringVarP(&lockServiceBaseURL, "lockservice-base-url", "", "http://lockservice.decap-system:2379", "Lock service base URL")
	unlockBuildCmd.Flags().StringVarP(&buildLockKey, "build-lock-key", "", "", "The build's build lock key.")

	putS3Cmd.Flags().StringVarP(&bucketName, "bucket-name", "", "", "S3 Bucket Name")
	putS3Cmd.Flags().StringVarP(&contentType, "content-type", "", "", "Content Type")
	putS3Cmd.Flags().StringVarP(&fileName, "filename", "", "", "File Name")

	recordBuildCmd.Flags().StringVarP(&tableName, "table-name", "", "", "DynamoDb build metadata table name")
	recordBuildCmd.Flags().StringVarP(&projectKey, "project-key", "", "", "Project key")
	recordBuildCmd.Flags().StringVarP(&branchToBuild, "branch", "", "", "Branch being built")
	recordBuildCmd.Flags().Int64VarP(&buildStartTime, "build-start-time", "", 0, "Unix time in seconds since the epoch when the build started")
	recordBuildCmd.Flags().Int64VarP(&buildResult, "build-result", "", 0, "Unix exit code of the executed build")
	recordBuildCmd.Flags().Int64VarP(&buildDuration, "build-duration", "", 0, "Duration of the build in seconds")

	BCToolCmd.PersistentFlags().StringVarP(&buildID, "build-id", "", "", "Build ID")
	BCToolCmd.PersistentFlags().StringVarP(&awsRegion, "aws-region", "", "", "AWS Region")
	BCToolCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Provide debug logging")

	BCToolCmd.AddCommand(versionCmd)
	BCToolCmd.AddCommand(unlockBuildCmd)
	BCToolCmd.AddCommand(putS3Cmd)
	BCToolCmd.AddCommand(recordBuildCmd)

	BCToolCmd.Execute()
}
