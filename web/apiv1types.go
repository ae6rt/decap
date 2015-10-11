package main

import "regexp"

type Meta struct {
	Error string `json:"error,omitempty"`
}

type Version struct {
	Meta
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	SDK     string `json:"sdk"`
}

type Projects struct {
	Meta
	Projects []Project `json:"projects"`
}

type Project struct {
	Team        string            `json:"team"`
	ProjectName string            `json:"project"`
	Descriptor  ProjectDescriptor `json:"descriptor,omitempty"`
	Sidecars    []string          `json:"sidecars,omitempty"`
}

type ProjectDescriptor struct {
	Image              string `json:"buildImage"`
	RepoManager        string `json:"repoManager"`
	RepoURL            string `json:"repoUrl"`
	RepoDescription    string `json:"repoDescription"`
	ManagedRefRegexStr string `json:"managedRefRegex"`
	regex              *regexp.Regexp
}

func (d ProjectDescriptor) isRefManaged(ref string) bool {
	return d.regex == nil || d.regex.MatchString(ref)
}

type Builds struct {
	Meta
	Builds []Build `json:"builds"`
}

type Build struct {
	ID         string `json:"id"`
	ProjectKey string `json:"projectKey"`
	Branch     string `json:"branch"`
	Result     int    `json:"result"`
	Duration   uint64 `json:"duration"`
	UnixTime   uint64 `json:"startTime"`
}

type Teams struct {
	Meta
	Teams []Team `json:"teams"`
}

type Team struct {
	Name string `json:"name"`
}

type Refs struct {
	Meta
	Refs []Ref `json:"refs"`
}

type Ref struct {
	RefID string `json:"ref"`
	Type  string `json:"type"`
}
