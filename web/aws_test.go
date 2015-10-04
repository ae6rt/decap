package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func TestDynamoDbGetBuilds(t *testing.T) {

	type F struct {
		AttrV struct {
			Key struct {
				S string `json:"S"`
			} `json:":pkey"`
			Since struct {
				N string `json:"N"`
			} `json:":since"`
		} `json:"ExpressionAttributeValues"`
		IndexName              string `json:"IndexName"`
		KeyConditionExpression string `json:"KeyConditionExpression"`
		Limit                  int    `json:"Limit"`
		ScanIndexForward       bool   `json:"ScanIndexForward"`
		TableName              string `json:"TableName"`
	}

	var v F
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &v)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(w, "")
	}))
	defer testServer.Close()

	config := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials("key", "secret", "")).WithRegion("region").WithMaxRetries(3).WithEndpoint(testServer.URL)
	c := AWSStorageService{config}

	_, err := c.GetBuildsByAtom(Atom{Team: "ae6rt", Library: "somelib"}, 0, 1)
	if err != nil {
		t.Fatal(err)
	}

	if v.AttrV.Key.S != "ae6rt/somelib" {
		t.Fatalf("Want ae6rt/somelib but got %s\n", v.AttrV.Key.S)
	}
	if v.AttrV.Since.N != "0" {
		t.Fatalf("Want 0 but got %s\n", v.AttrV.Since.N)
	}
	if v.IndexName != "project-key-build-start-time-index" {
		t.Fatalf("Want project-key-build-start-time-index but got %s\n", v.IndexName)
	}
	if v.KeyConditionExpression != "#pkey = :pkey and #bst > :since" {
		t.Fatalf("Want #pkey = :pkey and #bst > :since but got %s\n", v.KeyConditionExpression)
	}
	if v.Limit != 1 {
		t.Fatalf("Want 1 but got %d\n", v.Limit)
	}
	if v.ScanIndexForward {
		t.Fatal("Want false")
	}
	if v.TableName != "decap-build-metadata" {
		t.Fatal("Want decap-build-metadata but got %s\n", v.TableName)
	}
}
