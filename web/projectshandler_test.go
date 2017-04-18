package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
)

type ProjectHandlerProjectManager struct {
	ProjectManagerBaseMock
	allProjects map[string]v1.Project
}

func (t *ProjectHandlerProjectManager) GetProjects() map[string]v1.Project {
	return t.allProjects
}

func TestProjectsHandler(t *testing.T) {
	var tests = []struct {
		projects     map[string]v1.Project
		selectTeam   string
		wantProjects []v1.Project
	}{
		{
			projects: map[string]v1.Project{
				"ae6rt/p1": v1.Project{
					Team: "ae6rt",
				},
				"wn0owp/p2": v1.Project{
					Team: "wn0owp",
				},
			},
			selectTeam: "?team=",
			wantProjects: []v1.Project{
				v1.Project{
					Team: "ae6rt",
				},
				v1.Project{
					Team: "wn0owp",
				},
			},
		},
		{
			projects: map[string]v1.Project{
				"ae6rt/p1": v1.Project{
					Team: "ae6rt",
				},
				"wn0owp/p2": v1.Project{
					Team: "wn0owp",
				},
			},
			selectTeam: "?team=ae6rt",
			wantProjects: []v1.Project{
				v1.Project{
					Team: "ae6rt",
				},
			},
		},
	}

	for testNumber, test := range tests {
		req, _ := http.NewRequest("GET", "http://example.com/"+test.selectTeam, nil)

		w := httptest.NewRecorder()

		projectManager := &ProjectHandlerProjectManager{allProjects: test.projects}

		ProjectsHandler(projectManager)(w, req, nil)

		var got v1.Projects
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			t.Errorf("Test %d: unexepcted error: %v\n", testNumber, err)
		}

		if len(got.Projects) != len(test.wantProjects) {
			t.Errorf("Test %d: want %d, got %d\n", testNumber, len(test.projects), len(got.Projects))
		}

		for _, v := range got.Projects {
			if !containsProject(v, test.wantProjects) {
				t.Errorf("Test %d: should contain %v\n", testNumber, v)
			}
		}
	}
}

func containsProject(target v1.Project, expected []v1.Project) bool {
	for _, v := range expected {
		if target.Team == v.Team {
			return true
		}
	}
	return false
}
