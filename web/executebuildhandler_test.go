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

	atoms = map[string]Atom{
		"ae6rt/p1": Atom{
			Team: "ae6rt",
		},
		"wn0owp/p2": Atom{
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
	time.Sleep(500 * time.Millisecond)

	if mockDecap.event.Team() != "ae6rt" {
		t.Fatalf("Want ae6rt but got %s\n", mockDecap.event.Team())
	}
	if mockDecap.event.Project() != "p1" {
		t.Fatalf("Want p1 but got %s\n", mockDecap.event.Project())
	}

}

func TestExecuteBuildNoBranches(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com/ae6rt/p1", nil)
	if err != nil {
		log.Fatal(err)
	}

	atoms = map[string]Atom{
		"ae6rt/p1": Atom{
			Team: "ae6rt",
		},
		"wn0owp/p2": Atom{
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
	req, err := http.NewRequest("POST", "http://example.com/ae6rt/p1", nil)
	if err != nil {
		log.Fatal(err)
	}

	atoms = map[string]Atom{
		"ae6rt/p1": Atom{
			Team: "ae6rt",
		},
		"wn0owp/p2": Atom{
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
