package main

import "fmt"

func (e UserBuildEvent) Team() string {
	return e.Team_
}

func (e UserBuildEvent) Project() string {
	return e.Project_
}

func (e UserBuildEvent) Key() string {
	return fmt.Sprintf("%s/%s", e.Team_, e.Project_)
}

func (e UserBuildEvent) Refs() []string {
	return e.Refs_
}
