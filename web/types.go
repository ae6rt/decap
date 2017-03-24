package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/distrlocks"
	"github.com/ae6rt/decap/web/locks"
)

// TODO distinguish between pushes and branch creation.  Github has a header value that allows these to be differentiated.
// https://developer.github.com/webhooks/#delivery-headers
// https://gist.githubusercontent.com/ae6rt/53a25e726ac00b4cb535/raw/e3f412f6e7f408a56d0d691a1ec8b7658a495124/gh-create.json
// https://gist.githubusercontent.com/ae6rt/2be93f7d5edef8030b52/raw/29f591eb8ecc5555c55f1878b545613c1f9839b7/gh-push.json
type BuildEvent interface {
	Team() string
	Project() string
	Key() string
	Ref() string
	Hash() string
}

// DefaultBuilder models the main interface between Decap and Kubernetes.  This is the location where creating and deleting pods
// and locking or deferring builds takes place.
type DefaultBuilder struct {
	MasterURL       string
	UserName        string
	Password        string
	AWSAccessKeyID  string
	AWSAccessSecret string
	AWSRegion       string
	Locker          locks.Locker
	LockService     distrlocks.DistributedLockService
	DeferralService deferrals.DeferralService
	apiToken        string
	apiClient       *http.Client

	maxPods int

	buildScriptsRepo       string
	buildScriptsRepoBranch string

	tlsConfig *tls.Config

	logger *log.Logger
}

// RepoManagerCredential models the username and password for supported source code repository managers, such as Github or Atlassian Stash.
// For Github, the User is the OAuth2 access key and Password is the application's OAuth2 token.
type RepoManagerCredential struct {
	User     string
	Password string
}

// StorageService models the interaction between Decap and the persistent storage engine that stores build console logs, artifacts, and specific
// build metadata.
type StorageService interface {
	GetBuildsByProject(project v1.Project, sinceUnixTime uint64, limit uint64) ([]v1.Build, error)
	GetArtifacts(buildID string) ([]byte, error)
	GetConsoleLog(buildID string) ([]byte, error)
}

// Builder models the interaction between Decap and Kubernetes and the locking service that locks and defers builds.
type Builder interface {
	LaunchBuild(v1.UserBuildEvent) error
	CreatePod([]byte) error
	DeletePod(podName string) error
	DeferBuild(v1.UserBuildEvent) error
	LaunchDeferred(ticker <-chan time.Time)
	ClearDeferredBuild(key string) error
	DeferredBuilds() ([]v1.UserBuildEvent, error)
	PodWatcher()
}
