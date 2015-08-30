package main

import (
	"io/ioutil"
	"net/http"
)

type GithubEvent struct {
	PushEvent
}

func (stash GithubEvent) ProjectKey() string {
	// todo do something
	return ""
}

func (stash GithubEvent) Branches() []string {
	// todo do something
	branches := make([]string, 0)
	return branches
}

type GitHubHandler struct {
	K8sBase K8sBase
	Handler
}

func (han GitHubHandler) handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Println(err)
		return
	}
	Log.Printf("GitHub post-receive hook received: %s\n", data)
	go han.K8sBase.launchBuild(GithubEvent{})
}
