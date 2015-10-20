package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/locks"
	"github.com/julienschmidt/httprouter"
)

func TestGetDeferredBuilds(t *testing.T) {
	deferrals := []locks.Deferral{
		locks.Deferral{
			Data: `{"team": "t1", "project": "p1", "refs": ["master"]}`, // KEEP
			Key:  "/1",
		},
		locks.Deferral{
			Data: `{"team": "t1", "project": "p1", "refs": ["master"]}`, // dup of 0
			Key:  "/2",
		},
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/deferred", nil)

	locker := locks.NoOpLocker{Deferrals: deferrals}
	builder := DefaultBuilder{Locker: &locker}
	DeferredBuildsHandler(&builder)(w, req, []httprouter.Param{})
	if w.Code != 200 {
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}

	data, _ := ioutil.ReadAll(w.Body)
	var d v1.Deferred
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	if len(d.DeferredEvents) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(d.DeferredEvents))
	}
	if d.DeferredEvents[0].Hash() != "t1/p1/master" {
		t.Fatalf("Want t1/p1/master but got %d\n", d.DeferredEvents[0].Hash())
	}
}

func TestGetDeferredBuildsUnsupportedMethod(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "http://example.com/deferred", nil)

	mockBuilder := MockBuilder{}
	DeferredBuildsHandler(&mockBuilder)(w, req, []httprouter.Param{})
	if w.Code != 400 {
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
	data, _ := ioutil.ReadAll(w.Body)
	var d v1.Deferred
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	msg := d.Meta.Error
	if msg != "Unsupported method: PUT" {
		t.Fatalf("Expected Unsupported method: PUT but got %s\n", msg)
	}
}
