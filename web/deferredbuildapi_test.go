package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

type DeferredBuildsMock struct {
	MockDeferralService
	list []v1.UserBuildEvent
}

func (t DeferredBuildsMock) List() ([]v1.UserBuildEvent, error) {
	return t.list, nil
}

func TestGetDeferredBuilds(t *testing.T) {

	deferrals := []v1.UserBuildEvent{
		v1.UserBuildEvent{Team: "t1", Project: "p1", Ref: "master"},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/deferred", nil)

	builder := DefaultBuilder{DeferralService: DeferredBuildsMock{list: deferrals}}
	DeferredBuildsHandler(&builder)(w, req, []httprouter.Param{})
	if w.Code != 200 {
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}

	data, _ := ioutil.ReadAll(w.Body)
	var d []v1.UserBuildEvent
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	if len(d) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(d))
	}
	if d[0].Lockname() != "t1/p1/master" {
		t.Fatalf("Want t1/p1/master but got %s\n", d[0].Lockname())
	}
}

func TestClearDeferredBuild(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "http://example.com/deferred?key=/1", nil)

	mockBuilder := MockBuilder{}
	DeferredBuildsHandler(&mockBuilder)(w, req, []httprouter.Param{})
	if w.Code != 200 {
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}

	if mockBuilder.deferralKey != "/1" {
		t.Fatalf("Want /1 but got %s\n", mockBuilder.deferralKey)
	}
}

func TestClearDeferredBuildNoKey(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "http://example.com/deferred", nil)

	mockBuilder := MockBuilder{}
	DeferredBuildsHandler(&mockBuilder)(w, req, []httprouter.Param{})
	if w.Code != 400 {
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}

	data, _ := ioutil.ReadAll(w.Body)
	var d v1.UserBuildEvent
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	msg := d.Meta.Error
	if msg != "Missing or empty key parameter in clear deferred build" {
		t.Fatalf("Expected Missing or empty key parameter in clear deferred build but got %s\n", msg)
	}
}

func TestClearDeferredBuildError(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "http://example.com/deferred?key=/v1", nil)

	mockBuilder := MockBuilder{err: fmt.Errorf("boom")}
	DeferredBuildsHandler(&mockBuilder)(w, req, []httprouter.Param{})
	if w.Code != 500 {
		t.Fatalf("Want 500 but got %d\n", w.Code)
	}

	data, _ := ioutil.ReadAll(w.Body)
	var d v1.UserBuildEvent
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	msg := d.Meta.Error
	if msg != "boom" {
		t.Fatalf("Expected boom build but got %s\n", msg)
	}
}
