package main

import "strings"

// Captures a github post commit hook payload
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
