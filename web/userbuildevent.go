package main

import "fmt"

// UserBuildEvent captures a user-initiated build request.
type UserBuildEvent struct {
	TeamFld string
	LibraryFld string
	RefsFld []string
}

func (e UserBuildEvent) Team() string {
	return e.TeamFld
}

func (e UserBuildEvent) Library() string {
	return e.LibraryFld
}

func (e UserBuildEvent) ProjectKey() string {
	return fmt.Sprintf("%s/%s", e.TeamFld, e.LibraryFld)
}

func (e UserBuildEvent) Refs() []string {
	return e.RefsFld
}
