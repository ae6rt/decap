package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePod(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("wanted POST but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/api/v1/namespaces/decap/pods" {
			t.Fatalf("Want /api/v1/namespaces/decap/pods but got %s\n", url.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("Want application/json but got %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic YWRtaW46YWRtaW4xMjM=" { // base64(username:password)
			t.Fatalf("Want 'Basic YWRtaW46YWRtaW4xMjM=' but got '%s'\n", r.Header.Get("Authorization"))
		}
		w.WriteHeader(201)
		fmt.Fprint(w, "")
	}))
	defer testServer.Close()

	k8s := NewBuilder(testServer.URL, "admin", "admin123", "key", "sekrit", "us-west-1", NoOpLocker{}, "repo", "repobranch")
	err := k8s.CreatePod([]byte(""))
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
}
