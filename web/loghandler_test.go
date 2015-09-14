package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestLogHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	storageService := MockStorageService{data: []byte("foo")}

	w := httptest.NewRecorder()
	LogHandler(&storageService)(w, req, httprouter.Params{
		httprouter.Param{Key: "id", Value: "the-build-id"},
	},
	)

	if storageService.buildID != "the-build-id" {
		t.Fatalf("Want the-build-id but got %s\n", storageService.buildID)
	}

	if w.Header()["Content-Type"][0] != "application/x-gzip" {
		t.Fatalf("Want application/x-gzip but got %s\n", w.Header()["Content-Type"][0])
	}

	if w.Body.String() != "foo" {
		t.Fatalf("Want foo but got %s\n", w.Body.String())
	}
}

func TestLogHandlerWithError(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	storageService := MockStorageService{err: fmt.Errorf("boom")}

	w := httptest.NewRecorder()
	LogHandler(&storageService)(w, req, httprouter.Params{
		httprouter.Param{Key: "id", Value: "the-build-id"},
	},
	)

	if storageService.buildID != "the-build-id" {
		t.Fatalf("Want the-build-id but got %s\n", storageService.buildID)
	}

	if w.Code != 500 {
		t.Fatalf("Want 500 but got %d\n", w.Code)
	}
}
