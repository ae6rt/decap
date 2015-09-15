package main

import "fmt"

// UserBuildEvent captures a user-initiated build request.
type UserBuildEvent struct {
	team     string
	library  string
	branches []string
}

func (e UserBuildEvent) Team() string {
	return e.team
}

func (e UserBuildEvent) Library() string {
	return e.library
}

func (e UserBuildEvent) ProjectKey() string {
	return fmt.Sprintf("%s/%s", e.team, e.library)
}

func (e UserBuildEvent) Branches() []string {
	return e.branches
}
