package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	etcd "github.com/coreos/etcd/client"
)

type NoOpLocker struct {
	Locker
}

func (noop NoOpLocker) Lock(key, value string) (*etcd.Response, error) {
	return nil, nil
}

func (noop NoOpLocker) Unlock(key, value string) (*etcd.Response, error) {
	return nil, nil
}

func TestCreatePod(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("wanted POST but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/api/v1/namespaces/default/pods" {
			t.Fatalf("Want /api/v1/namespaces/default/pods but got %s\n", url.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("Want application/json but got %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Bearer thetoken" {
			t.Fatalf("Want 'Bearer thetoken' but got '%s'\n", r.Header.Get("Authorization"))
		}
		w.WriteHeader(201)
		fmt.Fprint(w, "")
	}))
	defer testServer.Close()

	k8s := NewK8s(testServer.URL, "thetoken", "admin", "admin123", NoOpLocker{})
	err := k8s.createPod([]byte(""))
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
}
