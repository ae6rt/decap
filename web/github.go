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

/*
// Team returns the Github owner
func (event GithubEvent) Team() string {
	return event.Repository.Owner.Name
}

// Project returns the Github repository name
func (event GithubEvent) Project() string {
	return event.Repository.Name
}

// Key returns the team + project hash map key
func (event GithubEvent) Key() string {
	return event.Repository.FullName
}

// Refs returns the references referenced in a Github push event
func (event GithubEvent) Ref() string {
	switch event.RefType {
	case "branch":
		return event.Ref
	case "tag":
		return ""
	default:
		return strings.ToLower(strings.Replace(event.Ref, "refs/heads/", "", -1))
	}
}
*/

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
		Team_:    event.Repository.Owner.Name,
		Project_: event.Repository.Name,
		Ref_:     refType,
	}
}

// TODO No recollection of why this is needed.  msp march 2017:  Hash is a simple hash function formed from the team name, project name, and references refs.
func (event GithubEvent) Hash() string {
	//	return fmt.Sprintf("%s/%s", event.Key(), strings.Join(event.Ref(), "/"))
	return "" // ??  best to break it to understand what it does
}
