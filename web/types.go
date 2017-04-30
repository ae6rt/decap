package main

import (
	"log"
	"time"

	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8sapi "k8s.io/client-go/pkg/api/v1"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
	"github.com/ae6rt/decap/web/projects"
)

// DefaultBuildManager models the main interface between Decap and Kubernetes.  This is the location where creating and deleting pods
// and locking or deferring builds takes place.
type DefaultBuildManager struct {
	lockService      lock.LockService
	deferralService  deferrals.DeferralService
	projectManager   projects.ProjectManager
	kubernetesClient k8sv1.PodsGetter
	maxPods          int
	logger           *log.Logger
}

// RepoManagerCredential models the username and password for supported source code repository managers, such as Github or Atlassian Stash.
// For Github, the User is the OAuth2 access key and Password is the application's OAuth2 token.
type RepoManagerCredential struct {
	User     string
	Password string
}

// BuildManager models the interaction between Decap and Kubernetes and the locking service that locks and defers builds.
type BuildManager interface {
	LaunchBuild(v1.UserBuildEvent) error
	CreatePod(*k8sapi.Pod) error
	DeletePod(podName string) error
	DeferBuild(deferrals.Deferrable) error
	LaunchDeferred(ticker <-chan time.Time)
	ClearDeferredBuild(key string) error
	DeferredBuilds() ([]deferrals.Deferrable, error)
	QueueIsOpen() bool
	CloseQueue()
	OpenQueue()
	PodWatcher()
}

// ClusterService models the Kubernetes client interface
// todo is this used?
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
