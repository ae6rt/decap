package main

import (
	"strings"

	"github.com/ae6rt/decap/web/api/v1"
)

// GithubEvent models a github post commit hook payload.
type GithubEvent struct {
	Ref        string           `json:"ref"`
	RefType    string           `json:"ref_type"`
	Repository GitHubRepository `json:"repository"`
}

// GitHubRepository models a Github repository.
type GitHubRepository struct {
	FullName string      `json:"full_name"`
	Name     string      `json:"name"`
	Owner    GithubOwner `json:"owner"`
}

// GithubOwner models the owner of a Github repository
type GithubOwner struct {
	Name string `json:"name"`
}

// BuildEvent turns a github event into a generic build event.
func (event GithubEvent) BuildEvent() v1.UserBuildEvent {
	var refType string

	switch event.RefType {
	case "branch":
		refType = event.Ref
	case "tag":
		refType = ""
	default:
		refType = strings.ToLower(strings.Replace(event.Ref, "refs/heads/", "", -1))
	}

	return v1.UserBuildEvent{
		Team:    event.Repository.Owner.Name,
		Project: event.Repository.Name,
		Ref:     refType,
	}
}
