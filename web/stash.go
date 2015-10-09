package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

// See https://confluence.atlassian.com/stash/post-service-webhook-for-stash-393284006.html for payload information.
type StashEvent struct {
	Repository StashRepository  `json:"repository"`
	RefChanges []StashRefChange `json:"refChanges"`

	BuildEvent
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

func (stash StashEvent) Team() string {
	return stash.Repository.Project.Key
}

func (stash StashEvent) Project() string {
	return stash.Repository.Slug
}

func (stash StashEvent) Key() string {
	return projectKey(stash.Repository.Project.Key, stash.Repository.Slug)
}

func (stash StashEvent) Refs() []string {
	branches := make([]string, 0)
	for _, v := range stash.RefChanges {
		branches = append(branches, strings.ToLower(strings.Replace(v.RefID, "refs/heads/", "", -1)))
	}
	return branches
}

func (event StashEvent) DeferralID() string {
	return ""
}

type StashHandler struct {
	decap Decap
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
	go handler.decap.LaunchBuild(event)
}
