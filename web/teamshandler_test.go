package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestTeamsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/teams", nil)
	if err != nil {
		log.Fatal(err)
	}

	projects = map[string]Atom{
		"ae6rt/p1": Atom{
			Team: "ae6rt",
		},
		"ae6rt/p2": Atom{
			Team: "ae6rt",
		},
		"wn0owp/p2": Atom{
			Team: "wn0owp",
		},
	}

	w := httptest.NewRecorder()
	TeamsHandler(w, req, httprouter.Params{})

	var teams Teams
	err = json.Unmarshal(w.Body.Bytes(), &teams)
	if err != nil {
		t.Fatal(err)
	}
	if len(teams.Teams) != 2 {
		t.Fatalf("Want 2 but got %d\n", len(teams.Teams))
	}
	for _, v := range teams.Teams {
		if !(v.Name == "ae6rt" || v.Name == "wn0owp") {
			t.Fatalf("Want ae6rt or wn0owp but got %d\n", v.Name)
		}
	}
}
