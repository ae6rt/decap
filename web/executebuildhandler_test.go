package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

func TestExecuteBuild(t *testing.T) {
	Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)

	req, err := http.NewRequest("POST", "http://example.com/ae6rt/p1?branch=master", nil)
	if err != nil {
		log.Fatal(err)
	}

	projectGetChan = make(chan map[string]v1.Project, 1)
	projectGetChan <- map[string]v1.Project{
		"ae6rt/p1": v1.Project{
			Team: "ae6rt",
		},
		"wn0owp/p2": v1.Project{
			Team: "wn0owp",
		},
	}

	w := httptest.NewRecorder()

	mockDecap := MockBuilder{}

	ExecuteBuildHandler(&mockDecap)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	// Let the goroutine finish.  Yuck.
	// TODO revisit this - March 2017
	time.Sleep(500 * time.Millisecond)

	if mockDecap.event.Team() != "ae6rt" {
		t.Fatalf("Want ae6rt but got %s\n", mockDecap.event.Team())
	}
	if mockDecap.event.Project() != "p1" {
		t.Fatalf("Want p1 but got %s\n", mockDecap.event.Project())
	}

}

func TestExecuteBuildNoBranches(t *testing.T) {
	Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)

	req, err := http.NewRequest("POST", "http://example.com/ae6rt/p1", nil)
	if err != nil {
		log.Fatal(err)
	}

	projectGetChan = make(chan map[string]v1.Project, 1)
	projectGetChan <- map[string]v1.Project{
		"ae6rt/p1": v1.Project{
			Team: "ae6rt",
		},
		"wn0owp/p2": v1.Project{
			Team: "wn0owp",
		},
	}

	w := httptest.NewRecorder()

	mockDecap := MockBuilder{}

	ExecuteBuildHandler(&mockDecap)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	if w.Code != 400 {
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
}

func TestExecuteBuildNoSuchProject(t *testing.T) {
	Log = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)

	req, err := http.NewRequest("POST", "http://example.com/ae6rt/p1", nil)
	if err != nil {
		log.Fatal(err)
	}

	projectGetChan = make(chan map[string]v1.Project, 1)
	projectGetChan <- map[string]v1.Project{
		"ae6rt/p1": v1.Project{
			Team: "ae6rt",
		},
		"wn0owp/p2": v1.Project{
			Team: "wn0owp",
		},
	}

	w := httptest.NewRecorder()

	mockDecap := MockBuilder{}

	ExecuteBuildHandler(&mockDecap)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "blah"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	if w.Code != 404 {
		t.Fatalf("Want 404 but got %d\n", w.Code)
	}
}
