package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func TestAWSS3Get(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
	}))
	defer testServer.Close()

	config := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials("key", "secret", "")).WithRegion("region").WithMaxRetries(3).WithEndpoint(testServer.URL).WithS3ForcePathStyle(true)

	c := AWSStorageService{config}
	data, err := c.GetArtifacts("buildID")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)

}
