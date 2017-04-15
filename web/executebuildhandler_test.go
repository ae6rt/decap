package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

type ExecManager struct {
	BuildManagerBaseMock
	wg           *sync.WaitGroup
	captureEvent v1.UserBuildEvent
	forceError   bool
}

func (t *ExecManager) LaunchBuild(event v1.UserBuildEvent) error {
	defer t.wg.Done()
	var err error
	if t.forceError {
		err = errors.New("Forced error")
	}
	t.captureEvent = event
	return err
}

func TestExecuteBuild(t *testing.T) {
	var tests = []struct {
		team             string
		project          string
		ref              string
		wantHTTPResponse int
	}{
		{
			team:             "ae6rt",
			project:          "p1",
			ref:              "?branch=master&branch=develop",
			wantHTTPResponse: 200,
		},
		{
			team:             "ae6rt",
			project:          "p1",
			wantHTTPResponse: 400,
		},
	}
	for testNumber, test := range tests {
		Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)

		req, err := http.NewRequest("POST", "http://example.com/api/v1/builds/"+test.team+"/"+test.project+test.ref, nil)
		if err != nil {
			log.Fatal(err)
		}

		w := httptest.NewRecorder()

		var wg sync.WaitGroup
		buildManager := ExecManager{wg: &wg}

		wg.Add(strings.Count(test.ref, "branch="))

		ExecuteBuildHandler(&buildManager, Log)(w, req, []httprouter.Param{httprouter.Param{Key: "team", Value: test.team}, httprouter.Param{Key: "project", Value: test.project}})

		if w.Code != test.wantHTTPResponse {
			t.Errorf("Test %d: want %d, got %d\n", testNumber, test.wantHTTPResponse, w.Code)
		}

		if w.Code != 200 {
			continue
		}

		wg.Wait()

		if buildManager.captureEvent.Team != test.team {
			t.Errorf("Test %d: Want %s but got %s\n", testNumber, test.team, buildManager.captureEvent.Team)
		}
		if buildManager.captureEvent.Project != test.project {
			t.Errorf("Test %d: Want %s but got %s\n", testNumber, test.project, buildManager.captureEvent.Project)
		}
	}
}
