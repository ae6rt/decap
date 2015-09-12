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

	projects = map[string]Project{
		"ae6rt/p1": Project{
			Team: "ae6rt",
		},
		"wn0owp/p2": Project{
			Team: "wn0owp",
		},
	}

	w := httptest.NewRecorder()
	ProjectsHandler(w, req, httprouter.Params{})

	var proj Projects
	err = json.Unmarshal(w.Body.Bytes(), &proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(proj.Projects) != 2 {
		t.Fatalf("Want 2 but got %d\n", len(proj.Projects))
	}
	for _, v := range proj.Projects {
		if !(v.Team == "ae6rt" || v.Team == "wn0owp") {
			t.Fatalf("Want ae6rt or wn0owp but got %d\n", v.Team)
		}
	}
}
