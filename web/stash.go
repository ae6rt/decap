package main

import (
	"fmt"
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