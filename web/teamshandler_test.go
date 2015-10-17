package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

func TestTeamsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/teams", nil)
	if err != nil {
		log.Fatal(err)
	}

	projectGetChan = make(chan map[string]v1.Project, 1)
	projectGetChan <- map[string]v1.Project{
		"ae6rt/p1": v1.Project{
			Team: "ae6rt",
		},
		"ae6rt/p2": v1.Project{
			Team: "ae6rt",
		},
		"wn0owp/p2": v1.Project{
			Team: "wn0owp",
		},
	}

	w := httptest.NewRecorder()
	TeamsHandler(w, req, httprouter.Params{})

	var teams v1.Teams
	err = json.Unmarshal(w.Body.Bytes(), &teams)
	if err != nil {
		t.Fatal(err)
	}
	if len(teams.Teams) != 2 {
		t.Fatalf("Want 2 but got %d\n", len(teams.Teams))
	}
	for _, v := range teams.Teams {
		if !(v.Name == "ae6rt" || v.Name == "wn0owp") {
			t.Fatalf("Want ae6rt or wn0owp but got %s\n", v.Name)
		}
	}
}
