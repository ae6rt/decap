package main

import (
	"log"
	"time"

	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8sapi "k8s.io/client-go/pkg/api/v1"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
)

// DefaultBuilder models the main interface between Decap and Kubernetes.  This is the location where creating and deleting pods
// and locking or deferring builds takes place.
type DefaultBuilder struct {
	lockService      lock.DistributedLockService
	deferralService  deferrals.DeferralService
	maxPods          int
	buildScripts     BuildScripts
	kubernetesClient k8sv1.PodsGetter
	logger           *log.Logger
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
	CreatePod(*k8sapi.Pod) error
	DeletePod(podName string) error
	DeferBuild(v1.UserBuildEvent) error
	LaunchDeferred(ticker <-chan time.Time)
	ClearDeferredBuild(key string) error
	DeferredBuilds() ([]v1.UserBuildEvent, error)
	PodWatcher()
}

// ClusterService models the Kubernetes client interface
type ClusterService interface {
}

// BuildScriptsRepo models where the build scripts are held
type BuildScripts struct {
	URL    string
	Branch string
}

// KubernetesClient is the subset we need of the full client API
type KubernetesClient interface {
	k8sv1.PodsGetter
	k8sv1.SecretsGetter
}

// AWSCredentials encapsulates the set of Decap AWS credentials
type AWSCredential struct {
	accessKey    string
	accessSecret string
	region       string
}
