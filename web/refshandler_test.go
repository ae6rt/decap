package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestProjectRefsNoSuchProject(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	projects = map[string]Atom{
		"ae6rt/p1": Atom{
			Team:       "ae6rt",
			Descriptor: AtomDescriptor{RepoManager: "github"},
		},
		"wn0owp/p2": Atom{
			Team: "wn0owp",
		},
	}

	githubClient := MockScmClient{}
	scmClients := map[string]SCMClient{"github": &githubClient}
	w := httptest.NewRecorder()
	ProjectRefsHandler(scmClients)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "nope"},
		httprouter.Param{Key: "library", Value: "p1"},
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

	projects = map[string]Atom{
		"ae6rt/p1": Atom{
			Team:       "ae6rt",
			Descriptor: AtomDescriptor{RepoManager: "subversion"},
		},
		"wn0owp/p2": Atom{
			Team: "wn0owp",
		},
	}

	githubClient := MockScmClient{}
	scmClients := map[string]SCMClient{"github": &githubClient}
	w := httptest.NewRecorder()
	ProjectRefsHandler(scmClients)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "library", Value: "p1"},
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

	projects = map[string]Atom{
		"ae6rt/p1": Atom{
			Team:       "ae6rt",
			Descriptor: AtomDescriptor{RepoManager: "github"},
		},
		"wn0owp/p2": Atom{
			Team: "wn0owp",
		},
	}

	githubClient := MockScmClient{branches: []Ref{Ref{RefID: "refs/heads/master"}}}
	scmClients := map[string]SCMClient{"github": &githubClient}
	w := httptest.NewRecorder()
	ProjectRefsHandler(scmClients)(w, req, httprouter.Params{
		httprouter.Param{Key: "team", Value: "ae6rt"},
		httprouter.Param{Key: "library", Value: "p1"},
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
