package app

import (
	"context"
	"errors"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/cluster"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
	"github.com/ae6rt/decap/web/projects"
	"github.com/ae6rt/decap/web/scmclients"
	"github.com/ae6rt/decap/web/storage"
)

var (
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
)

type Service interface {
	GetVersion(ctx context.Context) (v1.Version, error)
}

type DefaultService struct {
	version         v1.Version
	k8sClient       cluster.KubernetesClient
	deferralService deferrals.DeferralService
	storageService  storage.Service
	lockService     lock.LockService
	projectManager  projects.ProjectManager
	scmManagers     map[string]scmclients.SCMClient
}
