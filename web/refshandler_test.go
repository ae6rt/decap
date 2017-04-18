package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/scmclients"
	"github.com/julienschmidt/httprouter"
)

type RefsHandlerProjectManager struct {
	ProjectManagerBaseMock
	projects map[string]v1.Project
}

func (t *RefsHandlerProjectManager) GetProjectByTeamName(team, projectName string) (v1.Project, bool) {
	a, ok := t.projects[team+"/"+projectName]
	return a, ok
}

func TestProjectRefs(t *testing.T) {
	var tests = []struct {
		inTeam           string
		inProject        string
		projects         map[string]v1.Project
		wantHTTPResponse int
	}{
		{
			inTeam:    "nope",
			inProject: "p1",
			projects: map[string]v1.Project{
				"ae6rt/p1": v1.Project{
					Team:        "ae6rt",
					ProjectName: "p1",
					Descriptor:  v1.ProjectDescriptor{RepoManager: "github"},
				},
			},
			wantHTTPResponse: 404,
		},
		{
			inTeam:    "ae6rt",
			inProject: "p1",
			projects: map[string]v1.Project{
				"ae6rt/p1": v1.Project{
					Team:        "ae6rt",
					ProjectName: "p1",
					Descriptor: v1.ProjectDescriptor{
						RepoManager: "subversion",
					},
				},
			},
			wantHTTPResponse: 400,
		},
		{
			inTeam:    "ae6rt",
			inProject: "p1",
			projects: map[string]v1.Project{
				"ae6rt/p1": v1.Project{
					Team:        "ae6rt",
					ProjectName: "p1",
					Descriptor: v1.ProjectDescriptor{
						RepoManager: "github",
					},
				},
			},
			wantHTTPResponse: 200,
		},
	}

	for testNumber, test := range tests {
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		projectManager := &RefsHandlerProjectManager{projects: test.projects}

		scmClients := map[string]scmclients.SCMClient{"github": &scmclients.MockScmClient{}}

		w := httptest.NewRecorder()

		ProjectRefsHandler(projectManager, scmClients, log.New(ioutil.Discard, "", 0))(w, req, []httprouter.Param{httprouter.Param{Key: "team", Value: test.inTeam}, httprouter.Param{Key: "project", Value: test.inProject}})

		if w.Code != test.wantHTTPResponse {
			t.Errorf("Test %d: want %d but got %d\n", testNumber, test.wantHTTPResponse, w.Code)
		}
	}
}
