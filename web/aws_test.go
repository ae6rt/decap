package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func TestAWSS3GetArtifacts(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/decap-build-artifacts/buildID" {
			t.Fatalf("Want /decap-build-artifacts/buildID but got %s\n", r.URL.Path)
		}
		w.Write([]byte{0})
	}))
	defer testServer.Close()

	config := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials("key", "secret", "")).WithRegion("region").WithMaxRetries(3).WithEndpoint(testServer.URL).WithS3ForcePathStyle(true)

	c := AWSStorageService{config}
	data, err := c.GetArtifacts("buildID")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(data))
	}
	if data[0] != 0 {
		t.Fatalf("Want 0 but got %d\n", data[0])
	}
}

func TestAWSS3GetConsoleLogs(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/decap-console-logs/buildID" {
			t.Fatalf("Want /decap-console-logs/buildID but got %s\n", r.URL.Path)
		}
		w.Write([]byte{0})
	}))
	defer testServer.Close()

	config := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials("key", "secret", "")).WithRegion("region").WithMaxRetries(3).WithEndpoint(testServer.URL).WithS3ForcePathStyle(true)

	c := AWSStorageService{config}
	data, err := c.GetConsoleLog("buildID")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(data))
	}
	if data[0] != 0 {
		t.Fatalf("Want 0 but got %d\n", data[0])
	}
}
