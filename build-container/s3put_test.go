package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func TestS3Put(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			fmt.Printf("%s:%v\n", k, v)
		}
		if strings.ToUpper(r.Method) != "PUT" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/bucket/uuid" {
			t.Fatalf("Want /bucket/uuid but got %s\n", url.Path)
		}
		if r.Header.Get("Content-type") != "text/plain" {
			t.Fatalf("Want text/plain but got %s\n", r.Header.Get("Content-type"))
		}
		w.WriteHeader(200)
	}))
	defer testServer.Close()

	os.Setenv("AWS_ACCESS_KEY_ID ", "key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "sekrit")

	awsAccessKey = "key"
	awsAccessSecret = "sekrit"
	awsRegion = "us-west-1"

	awsConfig = func() *aws.Config {
		return aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(awsAccessKey, awsAccessSecret, "")).WithRegion(awsRegion).WithMaxRetries(3).
			WithEndpoint(testServer.URL).WithS3ForcePathStyle(true)
	}

	fileName = "/etc/hosts"
	bucketName = "bucket"
	contentType = "text/plain"
	buildID = "uuid"

	putS3Cmd.Run(nil, []string{})
}
