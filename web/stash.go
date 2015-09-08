package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// See https://confluence.atlassian.com/stash/post-service-webhook-for-stash-393284006.html for payload information.
type StashEvent struct {
	Repository StashRepository  `json:"repository"`
	RefChanges []StashRefChange `json:"refChanges"`

	PushEvent
}

type StashRepository struct {
	Slug    string       `json:"slug"`
	Project StashProject `json:"project"`
}

type StashProject struct {
	Key string `json:"key"`
}

type StashRefChange struct {
	RefID string `json:"refId"`
}

func (stash StashEvent) Parent() string {
	return stash.Repository.Project.Key
}

func (stash StashEvent) Library() string {
	return stash.Repository.Slug
}

func (stash StashEvent) ProjectKey() string {
	return fmt.Sprintf("%s/%s", stash.Repository.Project.Key, stash.Repository.Slug)
}

func (stash StashEvent) Branches() []string {
	branches := make([]string, 0)
	for _, v := range stash.RefChanges {
		branches = append(branches, strings.ToLower(strings.Replace(v.RefID, "refs/heads/", "", -1)))
	}
	return branches
}

type StashHandler struct {
	K8sBase DefaultDecap
}

func (handler StashHandler) handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Println(err)
		return
	}
	Log.Printf("Stash hook received: %s\n", data)

	var event StashEvent
	if err := json.Unmarshal(data, &event); err != nil {
		Log.Println(err)
		return
	}
	go handler.K8sBase.launchBuild(event)
}
