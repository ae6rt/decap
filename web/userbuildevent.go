package main

import (
	"fmt"
	"strings"
)

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

func (e UserBuildEvent) Hash() string {
	return fmt.Sprintf("%s/%s", e.Key(), strings.Join(e.Refs(), "/"))
}
