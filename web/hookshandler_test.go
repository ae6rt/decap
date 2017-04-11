package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/ziptools"
	"github.com/julienschmidt/httprouter"
)

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

	mockDecap := MockBuilder{}
	HooksHandler(BuildScripts{URL: "file://" + dir, Branch: "master"}, &mockDecap)(w, req, []httprouter.Param{httprouter.Param{Key: "repomanager", Value: "nosuchmanager"}})

	if w.Code != 400 {
		_ = os.RemoveAll(dir)
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
}

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

	mockDecap := MockBuilder{}
	HooksHandler(BuildScripts{URL: "file://" + dir, Branch: "master"}, &mockDecap)(w, req, []httprouter.Param{
		httprouter.Param{Key: "repomanager", Value: "buildscripts"},
	},
	)

	if w.Code != 200 {
		_ = os.RemoveAll(dir)
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}
}

func TestHooksHandlerGithub(t *testing.T) {
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
	req.Header.Set("X-Github-Event", "push")

	if err != nil {
		_ = os.RemoveAll(dir)
		log.Fatal(err)
	}

	w := httptest.NewRecorder()

	mockDecap := MockBuilder{}
	HooksHandler(BuildScripts{URL: "file://" + dir, Branch: "master"}, &mockDecap)(w, req, []httprouter.Param{
		httprouter.Param{Key: "repomanager", Value: "github"},
	},
	)

	if w.Code != 200 {
		_ = os.RemoveAll(dir)
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}

	// Wait for goroutine to run. Yuck.
	time.Sleep(1000 * time.Millisecond)

	if mockDecap.event.Team != "ae6rt" {
		t.Fatalf("Want ae6rt but got %s\n", mockDecap.event.Team)
	}
	if mockDecap.event.Project != "dynamodb-lab" {
		t.Fatalf("Want dynamodb-lab but got %s\n", mockDecap.event.Project)
	}
}

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

	mockDecap := MockBuilder{}
	HooksHandler(BuildScripts{URL: "file://" + dir, Branch: "master"}, &mockDecap)(w, req, []httprouter.Param{
		httprouter.Param{Key: "repomanager", Value: "github"},
	},
	)

}
