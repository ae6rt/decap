package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func TestDbPut(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		var x map[string]interface{}
		err = json.Unmarshal(data, &x)

		if x["TableName"].(string) != "table" {
			t.Fatalf("Want table but found %s\n", x["TableName"].(string))
		}

		item := x["Item"].(map[string]interface{})
		if item["branch"].(map[string]interface{})["S"].(string) != "branch" {
			t.Fatalf("Want branch but got %s\n", item["branch"].(map[string]interface{})["S"].(string) != "branch")
		}
		if item["build-duration"].(map[string]interface{})["N"].(string) != "3" {
			t.Fatalf("Want 3 but got %s\n", item["build-duration"].(map[string]interface{})["N"].(string))
		}
		if item["build-id"].(map[string]interface{})["S"].(string) != "uuid" {
			t.Fatalf("Want uuid but got %s\n", item["build-id"].(map[string]interface{})["S"].(string))
		}
		if item["build-result"].(map[string]interface{})["N"].(string) != "2" {
			t.Fatalf("Want 2 but got %s\n", item["build-result"].(map[string]interface{})["N"].(string))
		}
		if item["build-start-time"].(map[string]interface{})["N"].(string) != "1" {
			t.Fatalf("Want 1 but got %s\n", item["build-start-time"].(map[string]interface{})["N"].(string))
		}
		if item["project-key"].(map[string]interface{})["S"].(string) != "pkey" {
			t.Fatalf("Want pkey but got %s\n", item["project-key"].(map[string]interface{})["S"].(string) != "pkey")
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
