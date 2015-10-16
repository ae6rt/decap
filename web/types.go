package main

import "net/http"

// TODO distinguish between pushes and branch creation.  Github has a header value that allows these to be differentiated.
// https://developer.github.com/webhooks/#delivery-headers
// https://gist.githubusercontent.com/ae6rt/53a25e726ac00b4cb535/raw/e3f412f6e7f408a56d0d691a1ec8b7658a495124/gh-create.json
// https://gist.githubusercontent.com/ae6rt/2be93f7d5edef8030b52/raw/29f591eb8ecc5555c55f1878b545613c1f9839b7/gh-push.json
type BuildEvent interface {
	Team() string
	Project() string
	Key() string
	Refs() []string
	DeferralID() string
}

type DefaultBuilder struct {
	MasterURL       string
	UserName        string
	Password        string
	AWSAccessKeyID  string
	AWSAccessSecret string
	AWSRegion       string
	Locker          Locker

	apiToken  string
	apiClient *http.Client

	maxPods int

	buildScriptsRepo       string
	buildScriptsRepoBranch string
}

type RepoManagerCredential struct {
	User     string
	Password string
}

type StorageService interface {
	GetBuildsByProject(project Project, sinceUnixTime uint64, limit uint64) ([]Build, error)
	GetArtifacts(buildID string) ([]byte, error)
	GetConsoleLog(buildID string) ([]byte, error)
}

type Builder interface {
	LaunchBuild(buildEvent BuildEvent) error
	DeletePod(podName string) error
	DeferBuild(event BuildEvent, ref string) error
}

// UserBuildEvent captures a user-initiated build request.
type UserBuildEvent struct {
	Team_       string
	Project_    string
	Refs_       []string
	DeferralID_ string // when this object is unmarshaled from storage, set this field
}
