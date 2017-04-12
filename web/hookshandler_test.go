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

type HooksHandlerGithub struct {
	BaseLauncherMock
	wg           *sync.WaitGroup
	captureEvent v1.UserBuildEvent
	forceError   bool
}

func (t *HooksHandlerGithub) LaunchBuild(event v1.UserBuildEvent) error {
	defer t.wg.Done()

	var err error
	if t.forceError {
		err = errors.New("Forced error")
	}
	t.captureEvent = event
	return err
}

func TestHooksHandlerGithub(t *testing.T) {
	var tests = []struct {
		payload        string
		wantProjectKey string
		wantRef        string
	}{
		{

			payload: `{
  "ref": "refs/heads/master",
  "repository": {
    "id": 35129377,
    "name": "dynamodb-lab",
    "full_name": "ae6rt/dynamodb-lab",
    "owner": {
      "name": "ae6rt",
      "email": "ae6rt@users.noreply.github.com"
    }
  }
}`,
			wantProjectKey: "ae6rt/dynamodb-lab",
			wantRef:        "master",
		},
	}
	for testNumber, test := range tests {
		Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)

		req, _ := http.NewRequest("POST", "http://example.com/hooks", bytes.NewBufferString(test.payload))
		req.Header.Set("X-Github-Event", "push")

		w := httptest.NewRecorder()

		var wg sync.WaitGroup
		launcher := &HooksHandlerGithub{wg: &wg}

		wg.Add(1)
		HooksHandler(BuildScripts{}, launcher)(w, req, []httprouter.Param{httprouter.Param{Key: "repomanager", Value: "github"}})

		wg.Wait()

		if launcher.captureEvent.ProjectKey() != test.wantProjectKey {
			t.Errorf("Test %d: want %s, got %s\n", testNumber, test.wantProjectKey, launcher.captureEvent.ProjectKey())
		}

		if launcher.captureEvent.Ref != test.wantRef {
			t.Errorf("Test %d: want %s, got %s\n", testNumber, test.wantProjectKey, launcher.captureEvent.ProjectKey())
		}
	}
}

/*
func TestHooksHandlerNoRepoManager(t *testing.T) {
	Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)

	dir, err := ziptools.Unzip("testdata/buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	req, err := http.NewRequest("POST", "http://example.com/hooks/xxx", nil)
	if err != nil {
		_ = os.RemoveAll(dir)
		log.Fatal(err)
	}

	w := httptest.NewRecorder()

	mockDecap := BaseBuilderMock{}
	HooksHandler(BuildScripts{URL: "file://" + dir, Branch: "master"}, &mockDecap)(w, req, []httprouter.Param{httprouter.Param{Key: "repomanager", Value: "nosuchmanager"}})

	if w.Code != 400 {
		_ = os.RemoveAll(dir)
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
}
*/

/*
func TestHooksHandlerBuildScripts(t *testing.T) {
	Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)

	dir, err := ziptools.Unzip("testdata/buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	req, err := http.NewRequest("POST", "http://example.com/hooks/xxx", bytes.NewBufferString(""))
	if err != nil {
		_ = os.RemoveAll(dir)
		log.Fatal(err)
	}

	w := httptest.NewRecorder()

	projectSetChan = make(chan map[string]v1.Project, 1)

	mockDecap := BaseBuilderMock{}
	HooksHandler(BuildScripts{URL: "file://" + dir, Branch: "master"}, &mockDecap)(w, req, []httprouter.Param{
		httprouter.Param{Key: "repomanager", Value: "buildscripts"},
	},
	)

	if w.Code != 200 {
		_ = os.RemoveAll(dir)
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}
}
*/

/*
func TestHooksHandlerGithubNoEventTypeHeader(t *testing.T) {
	Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)

	dir, err := ziptools.Unzip("testdata/buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	req, err := http.NewRequest("POST", "http://example.com/hooks/xxx", bytes.NewBufferString(`
{
  "ref": "refs/heads/master",
  "repository": {
    "id": 35129377,
    "name": "dynamodb-lab",
    "full_name": "ae6rt/dynamodb-lab",
    "owner": {
      "name": "ae6rt",
      "email": "ae6rt@users.noreply.github.com"
    }
  }
}
`,
	))

	if err != nil {
		_ = os.RemoveAll(dir)
		log.Fatal(err)
	}

	w := httptest.NewRecorder()

	mockDecap := BaseBuilderMock{}
	HooksHandler(BuildScripts{URL: "file://" + dir, Branch: "master"}, &mockDecap)(w, req, []httprouter.Param{
		httprouter.Param{Key: "repomanager", Value: "github"},
	},
	)

}
*/
