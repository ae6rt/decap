package main

import (
	"bytes"
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
var buildStartTime string
var buildDuration string
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
			Log.Println("PUT successful")
		}
	},
}

/*
{
        "buildID": {
            "S": "$BUILD_ID"
        },
        "buildTime": {
            "N": "$START"
        },
        "projectKey": {
            "S": "$PROJECT_KEY"
        },
        "branch": {
            "S": "$BRANCH_TO_BUILD"
        },
        "isBuilding": {
            "N": "1"
        }
}
*/
var buildStartCmd = &cobra.Command{
	Use:   "build-start",
	Short: "mark a build as started in DynamoDb",
	Long:  "mark a build as started in DynamoDb",
	Run: func(cmd *cobra.Command, args []string) {

		config := aws.NewConfig().WithCredentials(credentials.NewEnvCredentials()).WithRegion(awsRegion).WithMaxRetries(3)
		svc := dynamodb.New(config)

		params := &dynamodb.PutItemInput{
			TableName: aws.String("decap-build-metadata"),
			Item:      map[string]*dynamodb.AttributeValue{"projectKey": {S: aws.String(":pkey")}},
		}
	},
}

func init() {
	putS3Cmd.Flags().StringVarP(&bucketName, "bucket-name", "b", "", "S3 Bucket Name")
	putS3Cmd.Flags().StringVarP(&buildID, "build-id", "i", "", "Build ID")
	putS3Cmd.Flags().StringVarP(&contentType, "content-type", "t", "", "Content Type")
	putS3Cmd.Flags().StringVarP(&fileName, "filename", "f", "", "File Name")
	putS3Cmd.Flags().StringVarP(&awsRegion, "aws-region", "r", "", "AWS Region")

	BCToolCmd.AddCommand(versionCmd)
	BCToolCmd.AddCommand(unlockBuildCmd)
	BCToolCmd.AddCommand(putS3Cmd)
}

func main() {
	BCToolCmd.Execute()
}
