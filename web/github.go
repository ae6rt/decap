package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type GithubEvent struct {
	Ref        string           `json:"ref"`
	Repository GitHubRepository `json:"repository"`
}

type GitHubRepository struct {
	FullName string      `json:"full_name"`
	Name     string      `json:"name"`
	Owner    GithubOwner `json:"owner"`
}

type GithubOwner struct {
	Name string `json:"name"`
}

func (event GithubEvent) Parent() string {
	return event.Repository.Owner.Name
}

func (event GithubEvent) Library() string {
	return event.Repository.Name
}

func (event GithubEvent) ProjectKey() string {
	return event.Repository.FullName
}

func (event GithubEvent) Branches() []string {
	return []string{strings.ToLower(strings.Replace(event.Ref, "refs/heads/", "", -1))}
}

type GitHubHandler struct {
	K8sBase DefaultDecap
	Handler
}

func (handler GitHubHandler) handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Println(err)
		return
	}

	var event GithubEvent
	if err := json.Unmarshal(data, &event); err != nil {
		Log.Println(err)
		return
	}
	Log.Printf("GitHub hook received: %s\n", data)
	go handler.K8sBase.launchBuild(event)
}
