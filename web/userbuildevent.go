package main

import "fmt"

func (e UserBuildEvent) Team() string {
	return e.TeamFld
}

func (e UserBuildEvent) Project() string {
	return e.ProjectFld
}

func (e UserBuildEvent) ProjectKey() string {
	return fmt.Sprintf("%s/%s", e.TeamFld, e.ProjectFld)
}

func (e UserBuildEvent) Refs() []string {
	return e.RefsFld
}
