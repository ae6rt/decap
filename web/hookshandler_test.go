package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

type HooksManager struct {
	BuildManagerBaseMock
	wg           *sync.WaitGroup
	captureEvent v1.UserBuildEvent
	forceError   bool
}

func (t *HooksManager) LaunchBuild(event v1.UserBuildEvent) error {
	defer t.wg.Done()

	var err error
	if t.forceError {
		err = errors.New("Forced error")
	}
	t.captureEvent = event
	return err
}

type HooksProjects struct {
	ProjectManagerBaseMock
	assembled bool
	set       bool
	get       bool
}

func (t *HooksProjects) Assemble() (map[string]v1.Project, error) {
	t.assembled = true
	return nil, nil
}

func (t *HooksProjects) Set(map[string]v1.Project) {
	t.set = true
}

func TestHooksHandler(t *testing.T) {
	var tests = []struct {
		endpoint             string
		HTTPHeaders          map[string][]string
		payload              string
		wantProjectKey       string
		wantRef              string
		wantHTTPResponseCode int
	}{
		{
			endpoint:             "github",
			HTTPHeaders:          map[string][]string{"X-Github-Event": []string{"create"}},
			payload:              `{"ref":"refs/heads/master","repository":{"name":"dynamodb-lab","full_name":"ae6rt/dynamodb-lab","owner":{"name":"ae6rt","email":"ae6rt@users.noreply.github.com"}}}`,
			wantProjectKey:       "ae6rt/dynamodb-lab",
			wantRef:              "master",
			wantHTTPResponseCode: 200,
		},
		{
			endpoint:             "github",
			HTTPHeaders:          map[string][]string{"X-Github-Event": []string{"push"}},
			payload:              `{"ref":"refs/heads/master","repository":{"name":"dynamodb-lab","full_name":"ae6rt/dynamodb-lab","owner":{"name":"ae6rt","email":"ae6rt@users.noreply.github.com"}}}`,
			wantProjectKey:       "ae6rt/dynamodb-lab",
			wantRef:              "master",
			wantHTTPResponseCode: 200,
		},
		{
			endpoint:             "github",
			HTTPHeaders:          map[string][]string{"X-Github-Event": []string{"<unsupported>"}},
			payload:              `{}`,
			wantHTTPResponseCode: 400,
		},
		{
			endpoint:             "buildscripts",
			payload:              `{}`,
			wantHTTPResponseCode: 200,
		},
	}

	for testNumber, test := range tests {
		Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)

		req, _ := http.NewRequest("POST", "http://example.com/hooks/"+test.endpoint, bytes.NewBufferString(test.payload))
		req.Header = test.HTTPHeaders

		w := httptest.NewRecorder()

		var wg sync.WaitGroup
		launcher := &HooksManager{wg: &wg}
		projectManager := &HooksProjects{}

		wg.Add(1)
		HooksHandler(projectManager, launcher, Log)(w, req, []httprouter.Param{httprouter.Param{Key: "repomanager", Value: test.endpoint}})

		if w.Code != test.wantHTTPResponseCode {
			t.Errorf("Test %d: want %d, got %d\n", testNumber, test.wantHTTPResponseCode, w.Code)
		}

		if test.wantHTTPResponseCode != 200 {
			continue
		}

		switch test.endpoint {
		case "buildscripts":
			if !projectManager.assembled {
				t.Errorf("Test %d:  expecting projectManager.Assemble() to have been called.\n", testNumber)
			}
			if !projectManager.set {
				t.Errorf("Test %d:  expecting projectManager.Set() to have been called.\n", testNumber)
			}
		default:
			wg.Wait()
			if launcher.captureEvent.ProjectKey() != test.wantProjectKey {
				t.Errorf("Test %d: want %s, got %s\n", testNumber, test.wantProjectKey, launcher.captureEvent.ProjectKey())
			}
			if launcher.captureEvent.Ref != test.wantRef {
				t.Errorf("Test %d: want %s, got %s\n", testNumber, test.wantProjectKey, launcher.captureEvent.ProjectKey())
			}
		}
	}
}
