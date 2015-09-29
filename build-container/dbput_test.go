package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

/*
{
    "Item": {
        "branch": {
            "S": "branch"
        },
        "build-duration": {
            "N": "3"
        },
        "build-id": {
            "S": "uuid"
        },
        "build-result": {
            "N": "2"
        },
        "build-start-time": {
            "N": "1"
        },
        "project-key": {
            "S": "pkey"
        }
    },
    "TableName": "table"
}
*/

func TestDbPut(t *testing.T) {

	type Bag struct {
		//		Item      Item   `json:"Item"`
		TableName string `json:"TableName"`
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(b))
		fmt.Println(r)
		for k, v := range r.Header {
			fmt.Printf("%s:%v\n", k, v)
		}
		if r.Method != "POST" {
			t.Fatalf("wanted POST but found %s\n", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer testServer.Close()

	awsAccessKey = "key"
	awsAccessSecret = "sekrit"
	awsRegion = "us-west-1"

	awsConfig = func() *aws.Config {
		return aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(awsAccessKey, awsAccessSecret, "")).WithRegion(awsRegion).WithMaxRetries(3).
			WithEndpoint(testServer.URL).WithS3ForcePathStyle(true)
	}

	tableName = "table"
	projectKey = "pkey"
	branchToBuild = "branch"
	buildID = "uuid"
	buildStartTime = 1
	buildResult = 2
	buildDuration = 3

	recordBuildCmd.Run(nil, nil)

}
