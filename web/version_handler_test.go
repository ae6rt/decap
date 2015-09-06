package main

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionHandler(t *testing.T) {
	buildInfo = "foo"
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	VersionHandler(w, req, httprouter.Params{})

	var version Version
	err = json.Unmarshal(w.Body.Bytes(), &version)
	if err != nil {
		t.Fatal(err)
	}

	if version.Version != "foo" {
		t.Fatalf("Want foo but got %s\n", version.Version)
	}
}
