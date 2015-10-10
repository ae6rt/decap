package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"
	"github.com/julienschmidt/httprouter"
)

func TestProjectRefsNoSuchProject(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	getThing = make(chan map[string]Project, 1)
	getThing <- map[string]Project{
		"ae6rt/p1": Project{
			Team:        "ae6rt",
			ProjectName: "p1",
			Descriptor:  ProjectDescriptor{RepoManager: "github"},
		},
		"wn0owp/p2": Project{
			Team: "wn0owp",
		},
	}
	fmt.Println("@@@ here test A1")

	scmClients := map[string]SCMClient{"github": &MockScmClient{}}

	w := httptest.NewRecorder()
	ProjectRefsHandler(scmClients)(w, req, []httprouter.Param{
		httprouter.Param{Key: "team", Value: "nope"},
		httprouter.Param{Key: "project", Value: "p1"},
	},
	)
	fmt.Println("@@@ here test A2")

	if w.Code != 404 {
		t.Fatalf("Want 404 but got %d\n", w.Code)
	}
}

func TestProjectRefsNoRepManager(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	getThing = make(chan map[string]Project, 1)
	getThing <- map[string]Project{
		"ae6rt/p1": Project{
			Team:        "ae6rt",
			ProjectName: "p1",
			Descriptor: ProjectDescriptor{
				RepoManager: "subversion",
			},
		},
		"wn0owp/p2": Project{
			Team:        "wn0owp",
			ProjectName: "p2",
		},
	}
	fmt.Println("@@@ here test B")

	githubClient := MockScmClient{}
	scmClients := map[string]SCMClient{"github": &githubClient}
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

	getThing = make(chan map[string]Project, 1)
	getThing <- map[string]Project{
		"ae6rt/p1": Project{
			Team:        "ae6rt",
			ProjectName: "p1",
			Descriptor: ProjectDescriptor{
				RepoManager: "github",
			},
		},
		"wn0owp/p2": Project{
			Team:        "wn0owp",
			ProjectName: "p2",
		},
	}
	fmt.Println("@@@ here test C")

	githubClient := MockScmClient{branches: []Ref{Ref{RefID: "refs/heads/master"}}}
	scmClients := map[string]SCMClient{"github": &githubClient}
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

	var b Refs
	json.Unmarshal(data, &b)
	if len(b.Refs) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(b.Refs))
	}
	if b.Refs[0].RefID != "refs/heads/master" {
		t.Fatalf("Want refs/heads/master but got %s\n", b.Refs[0].RefID)
	}
}
