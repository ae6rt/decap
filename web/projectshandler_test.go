package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestProjectsHandler(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com/teams", nil)
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
	ProjectsHandler(w, req, httprouter.Params{})

	var proj Atoms
	err = json.Unmarshal(w.Body.Bytes(), &proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(proj.Atoms) != 2 {
		t.Fatalf("Want 2 but got %d\n", len(proj.Atoms))
	}
	for _, v := range proj.Atoms {
		if !(v.Team == "ae6rt" || v.Team == "wn0owp") {
			t.Fatalf("Want ae6rt or wn0owp but got %s\n", v.Team)
		}
	}
}

func TestProjectsHandlerWithQuery(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com/teams?team=ae6rt", nil)
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
	ProjectsHandler(w, req, httprouter.Params{})

	var proj Atoms
	err = json.Unmarshal(w.Body.Bytes(), &proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(proj.Atoms) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(proj.Atoms))
	}

	expected := proj.Atoms[0]
	if expected.Team != "ae6rt" {
		t.Fatalf("Want ae6rt but got %s\n", expected.Team)
	}
}
