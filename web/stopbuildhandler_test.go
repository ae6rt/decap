package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

type StopBuildManager struct {
	BuildManagerBaseMock
	capturePodName string
	forceError     bool
}

func (t *StopBuildManager) DeletePod(s string) error {
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	t.capturePodName = s
	return err
}

func TestStopBuildHandler(t *testing.T) {
	var tests = []struct {
		buildID          string
		wantHTTPResponse int
		forceError       bool
	}{
		{
			buildID:          "id",
			wantHTTPResponse: 200,
		},
		{
			buildID:          "id",
			wantHTTPResponse: 500,
			forceError:       true,
		},
	}

	for testNumber, test := range tests {
		req, _ := http.NewRequest("DELETE", "http://example.com", nil)

		w := httptest.NewRecorder()

		buildManager := &StopBuildManager{forceError: test.forceError}
		StopBuildHandler(buildManager, log.New(ioutil.Discard, "", 0))(w, req, []httprouter.Param{httprouter.Param{Key: "id", Value: test.buildID}})

		if w.Code != test.wantHTTPResponse {
			t.Errorf("Test %d: want %d, got %d\n", testNumber, test.wantHTTPResponse, w.Code)
		}

		if w.Header().Get("Content-type") != "application/json" {
			t.Errorf("Test %d: want application/json but got %s\n", testNumber, w.Header().Get("Content-type"))
		}

		switch w.Code {
		case 200:
			if buildManager.capturePodName != test.buildID {
				t.Errorf("Test %d: want %s, got %s\n", testNumber, test.buildID, buildManager.capturePodName)
			}
		default:
		}
	}
}
