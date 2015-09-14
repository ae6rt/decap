package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestProjectBranchesNoSuchProject(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	projects = map[string]Project{
		"ae6rt/p1": Project{
			Team:       "ae6rt",
			Descriptor: ProjectDescriptor{RepoManager: "github"},
		},
		"wn0owp/p2": Project{
			Team: "wn0owp",
		},
	}

	githubClient := MockScmClient{}
	scmClients := map[string]SCMClient{"github": &githubClient}
	w := httptest.NewRecorder()
	ProjectBranchesHandler(scmClients)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "nope"},
		httprouter.Param{Key: "library", Value: "p1"},
	},
	)

	if w.Code != 404 {
		t.Fatalf("Want 404 but got %d\n", w.Code)
	}
}

func TestProjectBranchesNoRepManager(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	projects = map[string]Project{
		"ae6rt/p1": Project{
			Team:       "ae6rt",
			Descriptor: ProjectDescriptor{RepoManager: "subversion"},
		},
		"wn0owp/p2": Project{
			Team: "wn0owp",
		},
	}

	githubClient := MockScmClient{}
	scmClients := map[string]SCMClient{"github": &githubClient}
	w := httptest.NewRecorder()
	ProjectBranchesHandler(scmClients)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "library", Value: "p1"},
	},
	)

	if w.Code != 400 {
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
}
func TestProjectBranches(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	projects = map[string]Project{
		"ae6rt/p1": Project{
			Team:       "ae6rt",
			Descriptor: ProjectDescriptor{RepoManager: "github"},
		},
		"wn0owp/p2": Project{
			Team: "wn0owp",
		},
	}

	githubClient := MockScmClient{}
	scmClients := map[string]SCMClient{"github": &githubClient}
	w := httptest.NewRecorder()
	ProjectBranchesHandler(scmClients)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "library", Value: "p1"},
	},
	)

	if w.Code != 200 {
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}
}
