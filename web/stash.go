package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ae6rt/decap/web/api/v1"
)

// See https://confluence.atlassian.com/stash/post-service-webhook-for-stash-393284006.html for payload information.
// StashEvent models a post-commit hook payload from Atlassian Stash Git SCM.  It models the BuildEvent interface.
type StashEvent struct {
	Repository StashRepository  `json:"repository"`
	RefChanges []StashRefChange `json:"refChanges"`
}

// StashRepository models a Stash repository, consisting of a project and slug.  Slug is the foo part of scheme://host:port/project/foo.git Stash URL.
type StashRepository struct {
	Slug    string       `json:"slug"`
	Project StashProject `json:"project"`
}

// StashProject models the Project name for an Atlassian Stash repository.
type StashProject struct {
	Key string `json:"key"`
}

// StashRefChange models the branch or tag ref of a Stash post commit hook payload.
type StashRefChange struct {
	RefID string `json:"refId"`
}

// Team returns the Project-part of a Stash post commit hook payload.
func (stash StashEvent) Team() string {
	return stash.Repository.Project.Key
}

// Project returns the Slug part of a Stash post commit hook payload.
func (stash StashEvent) Project() string {
	return stash.Repository.Slug
}

// Key returns the project key / slug tuple.  This defines the key in a Decap map of projects that holds this project's configuration.
func (stash StashEvent) Key() string {
	// todo needs a rework
	///	return projectKey(stash.Repository.Project.Key, stash.Repository.Slug)
	return ""
}

// Refs returns the list of branches referenced in a Stash post commit hook payload.
func (stash StashEvent) Refs() []string {
	var branches []string
	for _, v := range stash.RefChanges {
		branches = append(branches, strings.ToLower(strings.Replace(v.RefID, "refs/heads/", "", -1)))
	}
	return branches
}

// StashHandler handles launching a build for Stash post commit hook events.
type StashHandler struct {
	decap BuildManager
}

// The http handler for handling Stash post commit hook events.
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

	for _, v := range event.Refs() {
		u := v1.UserBuildEvent{Team: event.Team(), Project: event.Project(), Ref: v}
		go func() {
			if err := handler.decap.LaunchBuild(u); err != nil {
				Log.Printf("Cannot launch build for event %+v: %v\n", event, err)
			}
		}()
	}
}
