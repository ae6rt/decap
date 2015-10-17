package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

func TestVersionHandler(t *testing.T) {
	buildVersion = "1"
	buildCommit = "abc"
	buildDate = "today"
	buildGoSDK = "1.5"

	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	VersionHandler(w, req, httprouter.Params{})

	var version v1.Version
	err = json.Unmarshal(w.Body.Bytes(), &version)
	if err != nil {
		t.Fatal(err)
	}

	if version.Version != "1" {
		t.Fatalf("Want 1 but got %s\n", version.Version)
	}
	if version.Commit != "abc" {
		t.Fatalf("Want abc but got %s\n", version.Commit)
	}
	if version.Date != "today" {
		t.Fatalf("Want today but got %s\n", version.Date)
	}
	if version.SDK != "1.5" {
		t.Fatalf("Want 1.5 but got %s\n", version.SDK)
	}
}
