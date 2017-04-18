package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

type TeamsHandlerProjectManager struct {
	ProjectManagerBaseMock
	projects map[string]v1.Project
}

func (t *TeamsHandlerProjectManager) GetProjects() map[string]v1.Project {
	return t.projects
}

func TestTeamsHandler(t *testing.T) {
	var tests = []struct {
		projects  map[string]v1.Project
		wantTeams []v1.Team
	}{
		{
			projects: map[string]v1.Project{
				"ae6rt/p1": v1.Project{
					Team: "ae6rt",
				},
				"ae6rt/p2": v1.Project{
					Team: "ae6rt",
				},
				"wn0owp/p2": v1.Project{
					Team: "wn0owp",
				},
			},
			wantTeams: []v1.Team{
				v1.Team{Name: "ae6rt"},
				v1.Team{Name: "wn0owp"},
			},
		},
	}
	for testNumber, test := range tests {
		req, _ := http.NewRequest("GET", "http://example.com/teams", nil)

		projectManager := &TeamsHandlerProjectManager{projects: test.projects}

		w := httptest.NewRecorder()
		TeamsHandler(projectManager)(w, req, httprouter.Params{})

		var teams v1.Teams
		err := json.Unmarshal(w.Body.Bytes(), &teams)
		if err != nil {
			t.Errorf("Test %d: unexpected error: %v\n", testNumber, err)
		}

		if len(teams.Teams) != len(test.wantTeams) {
			t.Errorf("Test %d: want %d, got %d\n", testNumber, len(test.projects), len(teams.Teams))
		}

		for _, v := range teams.Teams {
			if !(v.Name == "ae6rt" || v.Name == "wn0owp") {
				t.Errorf("Want ae6rt or wn0owp but got %s\n", v.Name)
			}
		}
	}
}
