package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
)

func TestExecuteBuild(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com/ae6rt/p1?branch=master", nil)
	if err != nil {
		log.Fatal(err)
	}

	projects = map[string]Project{
		"ae6rt/p1": Project{
			Team: "ae6rt",
		},
		"wn0owp/p2": Project{
			Team: "wn0owp",
		},
	}

	w := httptest.NewRecorder()

	mockDecap := MockDecap{}

	ExecuteBuildHandler(&mockDecap)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "library", Value: "p1"},
	},
	)

	// Let the goroutine finish.  Yuck.
	time.Sleep(500 * time.Millisecond)

	if mockDecap.event.Team() != "ae6rt" {
		t.Fatalf("Want ae6rt but got %s\n", mockDecap.event.Team())
	}
	if mockDecap.event.Library() != "p1" {
		t.Fatalf("Want p1 but got %s\n", mockDecap.event.Library())
	}

}

func TestExecuteBuildNoBranches(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com/ae6rt/p1", nil)
	if err != nil {
		log.Fatal(err)
	}

	projects = map[string]Project{
		"ae6rt/p1": Project{
			Team: "ae6rt",
		},
		"wn0owp/p2": Project{
			Team: "wn0owp",
		},
	}

	w := httptest.NewRecorder()

	mockDecap := MockDecap{}

	ExecuteBuildHandler(&mockDecap)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "library", Value: "p1"},
	},
	)

	if w.Code != 400 {
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
}

func TestExecuteBuildNoSuchProject(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com/ae6rt/p1", nil)
	if err != nil {
		log.Fatal(err)
	}

	projects = map[string]Project{
		"ae6rt/p1": Project{
			Team: "ae6rt",
		},
		"wn0owp/p2": Project{
			Team: "wn0owp",
		},
	}

	w := httptest.NewRecorder()

	mockDecap := MockDecap{}

	ExecuteBuildHandler(&mockDecap)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "blah"},
		httprouter.Param{Key: "library", Value: "p1"},
	},
	)

	if w.Code != 404 {
		t.Fatalf("Want 404 but got %d\n", w.Code)
	}
}