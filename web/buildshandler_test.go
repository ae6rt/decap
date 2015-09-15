package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestBuildsHandlerSinceNotUnsigned(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com?since=-1", nil)
	if err != nil {
		log.Fatal(err)
	}

	storageService := MockStorageService{}

	w := httptest.NewRecorder()
	BuildsHandler(&storageService)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "library", Value: "p1"},
	},
	)

	var b Builds
	err = json.Unmarshal(w.Body.Bytes(), &b)
	if err != nil {
		t.Fatal(err)
	}
	if b.Error == "" {
		t.Fatal("Expected an error")
	}
}
