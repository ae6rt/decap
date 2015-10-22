package main

import (
	"fmt"
	"strings"
)

// GithubEvent models a github post commit hook payload
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
func (event GithubEvent) Refs() []string {
	switch event.RefType {
	case "branch":
		return []string{event.Ref}
	case "tag":
		return []string{}
	default:
		return []string{strings.ToLower(strings.Replace(event.Ref, "refs/heads/", "", -1))}
	}
}

// Hash is a simple hash function formed from the team name, project name, and references refs.
func (event GithubEvent) Hash() string {
	return fmt.Sprintf("%s/%s", event.Key(), strings.Join(event.Refs(), "/"))
}
