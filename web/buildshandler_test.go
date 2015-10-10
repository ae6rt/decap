package main

import (
	"encoding/json"
	"fmt"
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
	BuildsHandler(&storageService)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	var b Builds
	err = json.Unmarshal(w.Body.Bytes(), &b)
	if err != nil {
		t.Fatal(err)
	}
	if w.Code != 400 {
		t.Fatalf("Expected 400 but got %d\n", w.Code)
	}
	if b.Error == "" {
		t.Fatal("Expected an error because since is signed")
	}
}

func TestBuildsHandlerLimitNotUnsigned(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com?limit=-1", nil)
	if err != nil {
		log.Fatal(err)
	}

	storageService := MockStorageService{}
	w := httptest.NewRecorder()
	BuildsHandler(&storageService)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	var b Builds
	err = json.Unmarshal(w.Body.Bytes(), &b)
	if err != nil {
		t.Fatal(err)
	}
	if w.Code != 400 {
		t.Fatalf("Expected 400 but got %d\n", w.Code)
	}
	if b.Error == "" {
		t.Fatal("Expected an error because limit is signed")
	}
}

func TestBuildsHandlerWithStorageServiceError(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com?since=1&limit=2", nil)
	if err != nil {
		log.Fatal(err)
	}

	storageService := MockStorageService{err: fmt.Errorf("boom")}
	w := httptest.NewRecorder()
	BuildsHandler(&storageService)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	var b Builds
	err = json.Unmarshal(w.Body.Bytes(), &b)
	if err != nil {
		t.Fatal(err)
	}
	if w.Code != 502 {
		t.Fatalf("Expected 502 but got %d\n", w.Code)
	}
	if b.Error == "" {
		t.Fatal("Expected an error")
	}
	if storageService.project.Team != "ae6rt" {
		t.Fatalf("Want ae6rt but got %s\n", storageService.project.Team)
	}
	if storageService.project.ProjectName != "p1" {
		t.Fatalf("Want p1 but got %s\n", storageService.project.ProjectName)
	}
	if storageService.sinceUnixTime != 1 {
		t.Fatalf("Want 1 but got %d\n", storageService.sinceUnixTime)
	}
	if storageService.limit != 2 {
		t.Fatalf("Want 2 but got %d\n", storageService.limit)
	}
}

func TestBuildsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com?since=1&limit=2", nil)
	if err != nil {
		log.Fatal(err)
	}

	storageService := MockStorageService{builds: []Build{Build{ID: "the-id"}}}
	w := httptest.NewRecorder()
	BuildsHandler(&storageService)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	var b Builds
	err = json.Unmarshal(w.Body.Bytes(), &b)
	if err != nil {
		t.Fatal(err)
	}
	if b.Error != "" {
		t.Fatalf("Unexpected an error: %s\n", b.Error)
	}
	if storageService.project.Team != "ae6rt" {
		t.Fatalf("Want ae6rt but got %s\n", storageService.project.Team)
	}
	if storageService.project.ProjectName != "p1" {
		t.Fatalf("Want p1 but got %s\n", storageService.project.ProjectName)
	}
	if storageService.sinceUnixTime != 1 {
		t.Fatalf("Want 1 but got %d\n", storageService.sinceUnixTime)
	}
	if storageService.limit != 2 {
		t.Fatalf("Want 2 but got %d\n", storageService.limit)
	}
	if len(b.Builds) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(b.Builds))
	}
	if b.Builds[0].ID != "the-id" {
		t.Fatalf("Want the-id but got %s\n", b.Builds[0].ID)
	}
}
