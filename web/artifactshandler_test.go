package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestArtifactsHandler(t *testing.T) {
	Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadFile("testdata/sample.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	storageService := MockStorageService{data: data}

	w := httptest.NewRecorder()
	ArtifactsHandler(&storageService)(w, req, []httprouter.Param{
		httprouter.Param{Key: "id", Value: "the-build-id"},
	},
	)

	if storageService.buildID != "the-build-id" {
		t.Fatalf("Want the-build-id but got %s\n", storageService.buildID)
	}

	if w.Header()["Content-Type"][0] != "application/x-gzip" {
		t.Fatalf("Want application/x-gzip but got %s\n", w.Header()["Content-Type"][0])
	}

	if len(w.Body.Bytes()) != 7037 {
		t.Fatalf("Want 7037 but got %d\n", len(w.Body.Bytes()))
	}
}

func TestArtifactsHandlerManifestOnly(t *testing.T) {
	Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/plain")

	data, err := ioutil.ReadFile("testdata/sample.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	storageService := MockStorageService{data: data}

	w := httptest.NewRecorder()
	ArtifactsHandler(&storageService)(w, req, []httprouter.Param{
		httprouter.Param{Key: "id", Value: "the-build-id"},
	},
	)

	if storageService.buildID != "the-build-id" {
		t.Fatalf("Want the-build-id but got %s\n", storageService.buildID)
	}

	if w.Header()["Content-Type"][0] != "text/plain" {
		t.Fatalf("Want text/plain but got %s\n", w.Header()["Content-Type"][0])
	}

	if !strings.Contains(w.Body.String(), "artifactshandler_test.go") {
		t.Fatalf("Want contains artifactshandler_test.go: %s\n", w.Body.String())
	}
}

func TestArtifactsHandlerWithError(t *testing.T) {
	Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)
	req, err := http.NewRequest("POST", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	storageService := MockStorageService{err: fmt.Errorf("boom")}

	w := httptest.NewRecorder()
	ArtifactsHandler(&storageService)(w, req, []httprouter.Param{
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
