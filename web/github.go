package main

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Captures a github post commit hook payload
type GithubEvent struct {
	Ref        string           `json:"ref"`
	RefType    string           `json:"ref_type"`
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

func (event GithubEvent) Team() string {
	return event.Repository.Owner.Name
}

func (event GithubEvent) Project() string {
	return event.Repository.Name
}

func (event GithubEvent) Key() string {
	return event.Repository.FullName
}

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

func (event GithubEvent) DeferralID() string {
	s := fmt.Sprintf("%s/%s", event.Key(), strings.Join(event.Refs(), "/"))
	return hex.EncodeToString([]byte(s))
}
