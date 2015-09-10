package main

import "fmt"

// UserBuildEvent captures a user-initiated build request.
type UserBuildEvent struct {
	parent   string
	library  string
	branches []string
}

func (e UserBuildEvent) Parent() string {
	return e.parent
}

func (e UserBuildEvent) Library() string {
	return e.library
}

func (e UserBuildEvent) ProjectKey() string {
	return fmt.Sprintf("%s/%s", e.parent, e.library)
}

func (e UserBuildEvent) Branches() []string {
	return e.branches
}
