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

type DeferredBuildManagerMock struct {
	BuildManagerBaseMock
	list       []v1.UserBuildEvent
	captureKey string
	forceError bool
}

func (t *DeferredBuildManagerMock) DeferredBuilds() ([]v1.UserBuildEvent, error) {
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return t.list, err
}

func (t *DeferredBuildManagerMock) ClearDeferredBuild(key string) error {
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	t.captureKey = key
	return err
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

		buildManager := &DeferredBuildManagerMock{list: test.deferrals, forceError: test.forceError}

		DeferredBuildsHandler(buildManager)(w, req, []httprouter.Param{})

		if w.Code != test.wantHTTPResponse {
			t.Errorf("Test %d: want %d but got %d\n", testNumber, test.wantHTTPResponse, w.Code)
		}

		if w.Header().Get("Content-type") != "application/json" {
			t.Errorf("Test %d: want application/json but got %s\n", testNumber, w.Header().Get("Content-type"))
		}

		switch w.Code {
		case 200:
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
		default:
		}
	}
}

func TestClearDeferredBuild(t *testing.T) {
	var tests = []struct {
		key              string
		wantHTTPResponse int
		forceError       bool
	}{
		{
			key:              "abc",
			wantHTTPResponse: 200,
		},
		{
			key:              "",
			wantHTTPResponse: 400,
		},
		{
			key:              "abc",
			wantHTTPResponse: 500,
			forceError:       true,
		},
	}

	for testNumber, test := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "http://example.com/deferred?key="+test.key, nil)

		buildManager := &DeferredBuildManagerMock{forceError: test.forceError}
		DeferredBuildsHandler(buildManager)(w, req, []httprouter.Param{})

		if w.Code != test.wantHTTPResponse {
			t.Errorf("Test %d: want %d but got %d\n", testNumber, test.wantHTTPResponse, w.Code)
		}

		if w.Header().Get("Content-type") != "application/json" {
			t.Errorf("Test %d: want application/json but got %s\n", testNumber, w.Header().Get("Content-type"))
		}

		switch w.Code {
		case 200:
			if buildManager.captureKey != test.key {
				t.Errorf("Test %d: want %s, got %s\n", testNumber, test.key, buildManager.captureKey)
			}
		case 400:
		case 500:
		default:
		}
	}
}
