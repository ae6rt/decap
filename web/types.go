package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
)

// DefaultBuilder models the main interface between Decap and Kubernetes.  This is the location where creating and deleting pods
// and locking or deferring builds takes place.
type DefaultBuilder struct {
	masterURL       string
	masterUsername  string
	masterPassword  string
	awsAccessKeyID  string
	awsAccessSecret string
	awsRegion       string
	lockService     lock.DistributedLockService
	deferralService deferrals.DeferralService
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
