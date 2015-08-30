package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type StashContainer struct {
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

func (stash StashContainer) ProjectKey() string {
	return fmt.Sprintf("%s/%s", stash.Repository.Project, stash.Repository.Slug)
}

func (stash StashContainer) Branches() []string {
	branches := make([]string, 0)
	for _, v := range stash.RefChanges {
		branches = append(branches, strings.ToLower(strings.Replace(v.RefID, "refs/heads/", "", -1)))
	}
	return branches
}

type StashHandler struct {
	K8sBase K8sBase
	Handler
}

func (han StashHandler) handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Println(err)
		return
	}
	Log.Printf("post-receive hook received: %s\n", data)

	var stashContainer StashContainer
	if err := json.Unmarshal(data, &stashContainer); err != nil {
		Log.Println(err)
		return
	}
	go han.K8sBase.build(stashContainer)
}
