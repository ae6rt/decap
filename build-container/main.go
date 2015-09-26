package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ae6rt/decap/build-container/locks"
	"github.com/spf13/cobra"
)

var buildInfo string

var bucketName string
var buildID string
var contentType string
var fileName string

var Log *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

var BCToolCmd = &cobra.Command{
	Use:   "bctool",
	Short: "bctool is a multifunction build container tool.",
	Long:  `A multifunction build container tool that unlocks builds, uploads files to S3, and puts items to DynamoDb`,
	Run: func(cmd *cobra.Command, args []string) {
	},
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
		fmt.Println("Print: " + strings.Join(args, " "))
	},
}

var buildStartCmd = &cobra.Command{
	Use:   "build-start",
	Short: "mark a build as started in DynamoDb",
	Long:  "mark a build as started in DynamoDb",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Print: " + strings.Join(args, " "))
	},
}

func main() {
	putS3Cmd.Flags().StringVarP(&bucketName, "bucket-name", "b", "", "S3 Bucket Name")
	putS3Cmd.Flags().StringVarP(&buildID, "build-id", "i", "", "Build ID")
	putS3Cmd.Flags().StringVarP(&contentType, "content-type", "t", "", "Content Type")
	putS3Cmd.Flags().StringVarP(&fileName, "filename", "f", "", "File Name")

	BCToolCmd.AddCommand(versionCmd)
	BCToolCmd.AddCommand(unlockBuildCmd)
	BCToolCmd.AddCommand(putS3Cmd)
	BCToolCmd.Execute()
}
