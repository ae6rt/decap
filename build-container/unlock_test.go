package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnlock(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("wanted DELETE but found %s\n", r.Method)
		}
		url := r.URL
		if url.Path != "/v2/keys/buildlocks/lockit" {
			t.Fatalf("Want /v2/keys/buildlocks but got %s\n", url.Path)
		}
		if url.Query().Get("prevValue") != "uuid" {
			t.Fatalf("Want uuid but got %s\n", url.Query().Get("uuid"))
		}
		w.WriteHeader(201)
	}))
	defer testServer.Close()

	lockServiceBaseURL = testServer.URL
	buildLockKey = "lockit"
	buildID = "uuid"

	unlockBuildCmd.Run(nil, []string{})
}
