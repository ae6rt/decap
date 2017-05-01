package buildmanager

import (
	"time"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/cluster"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
	"github.com/ae6rt/decap/web/projects"
	log "github.com/go-kit/kit/log"
	k8sapi "k8s.io/client-go/pkg/api/v1"
)

// DefaultBuildManager models the main interface between Decap and Kubernetes.  This is the location where creating and deleting pods
// and locking or deferring builds takes place.
type DefaultBuildManager struct {
	lockService      lock.LockService
	deferralService  deferrals.DeferralService
	projectManager   projects.ProjectManager
	kubernetesClient cluster.KubernetesClient
	maxPods          int
	logger           log.Logger
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
