package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/scmclients"
	"github.com/julienschmidt/httprouter"
)

func TestProjectRefsNoSuchProject(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	projectGetChan = make(chan map[string]v1.Project, 1)
	projectGetChan <- map[string]v1.Project{
		"ae6rt/p1": v1.Project{
			Team:        "ae6rt",
			ProjectName: "p1",
			Descriptor:  v1.ProjectDescriptor{RepoManager: "github"},
		},
		"wn0owp/p2": v1.Project{
			Team: "wn0owp",
		},
	}

	scmClients := map[string]scmclients.SCMClient{"github": &scmclients.MockScmClient{}}

	w := httptest.NewRecorder()
	ProjectRefsHandler(scmClients)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "nope"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	if w.Code != 404 {
		t.Fatalf("Want 404 but got %d\n", w.Code)
	}
}

func TestProjectRefsNoRepManager(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	projectGetChan = make(chan map[string]v1.Project, 1)
	projectGetChan <- map[string]v1.Project{
		"ae6rt/p1": v1.Project{
			Team:        "ae6rt",
			ProjectName: "p1",
			Descriptor: v1.ProjectDescriptor{
				RepoManager: "subversion",
			},
		},
		"wn0owp/p2": v1.Project{
			Team:        "wn0owp",
			ProjectName: "p2",
		},
	}

	githubClient := scmclients.MockScmClient{}
	scmClients := map[string]scmclients.SCMClient{"github": &githubClient}
	w := httptest.NewRecorder()
	ProjectRefsHandler(scmClients)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	if w.Code != 400 {
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
}

func TestProjectRefsGithub(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	projectGetChan = make(chan map[string]v1.Project, 1)
	projectGetChan <- map[string]v1.Project{
		"ae6rt/p1": v1.Project{
			Team:        "ae6rt",
			ProjectName: "p1",
			Descriptor: v1.ProjectDescriptor{
				RepoManager: "github",
			},
		},
		"wn0owp/p2": v1.Project{
			Team:        "wn0owp",
			ProjectName: "p2",
		},
	}

	githubClient := scmclients.MockScmClient{Branches: []v1.Ref{v1.Ref{RefID: "refs/heads/master"}}}
	scmClients := map[string]scmclients.SCMClient{"github": &githubClient}
	w := httptest.NewRecorder()
	ProjectRefsHandler(scmClients)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)

	if w.Code != 200 {
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}

	data := w.Body.Bytes()

	var b v1.Refs
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
	if len(b.Refs) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(b.Refs))
	}
	if b.Refs[0].RefID != "refs/heads/master" {
		t.Fatalf("Want refs/heads/master but got %s\n", b.Refs[0].RefID)
	}
}
