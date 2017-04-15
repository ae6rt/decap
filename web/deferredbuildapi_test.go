package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

type DeferredBuildsMock struct {
	MockDeferralService
	list       []v1.UserBuildEvent
	forceError bool
}

func (t DeferredBuildsMock) List() ([]v1.UserBuildEvent, error) {
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return t.list, err
}

func TestGetDeferredBuilds(t *testing.T) {
	var tests = []struct {
		deferrals        []v1.UserBuildEvent
		forceError       bool
		wantHTTPResponse int
	}{
		{
			deferrals: []v1.UserBuildEvent{
				v1.UserBuildEvent{Team: "t1", Project: "p1", Ref: "master"},
				v1.UserBuildEvent{Team: "t2", Project: "p8", Ref: "master"},
			},
			wantHTTPResponse: 200,
		},
		{
			forceError:       true,
			wantHTTPResponse: 500,
		},
	}

	for testNumber, test := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "http://example.com/deferred", nil)

		deferralService := DeferredBuildsMock{list: test.deferrals, forceError: test.forceError}
		buildManager := &DefaultBuildManager{deferralService: deferralService}

		DeferredBuildsHandler(buildManager)(w, req, []httprouter.Param{})
		if w.Code != test.wantHTTPResponse {
			t.Errorf("Test %d: want %d but got %d\n", testNumber, test.wantHTTPResponse, w.Code)
		}

		if w.Header().Get("Content-type") != "application/json" {
			t.Errorf("Test %d: want application/json but got %s\n", testNumber, w.Header().Get("Content-type"))
		}

		if w.Code != 200 {
			continue
		}

		data, _ := ioutil.ReadAll(w.Body)
		var gotEvents []v1.UserBuildEvent
		if err := json.Unmarshal(data, &gotEvents); err != nil {
			t.Errorf("Unexpected error: %v\n", err)
		}

		if len(gotEvents) != len(test.deferrals) {
			t.Errorf("Test %d: want %d but got %d\n", testNumber, len(test.deferrals), len(gotEvents))
		}

		for k, v := range test.deferrals {
			if gotEvents[k].Lockname() != v.Lockname() {
				t.Errorf("Test %d: want %s but got %s\n", testNumber, v.Lockname(), gotEvents[k].Lockname())
			}
		}
	}
}

/*
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

*/
