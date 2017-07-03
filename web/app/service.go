package app

import (
	"context"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/cluster"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
	"github.com/ae6rt/decap/web/projects"
	"github.com/ae6rt/decap/web/scmclients"
	"github.com/ae6rt/decap/web/storage"

	kitlog "github.com/go-kit/kit/log"
)

func New(version v1.Version, k8sClient cluster.KubernetesClient, deferralService deferrals.DeferralService, storageService storage.Service, lockService lock.LockService,
	projectManager projects.ProjectManager, scmManagers map[string]scmclients.SCMClient, logger kitlog.Logger) Service {
	return DefaultService{
		version:         version,
		k8sClient:       k8sClient,
		deferralService: deferralService,
		storageService:  storageService,
		lockService:     lockService,
		projectManager:  projectManager,
		scmManagers:     scmManagers,
		logger:          logger,
	}
}

func (t DefaultService) GetVersion(ctx context.Context) (v1.Version, error) {
	return t.version, nil
}

func NewService(version v1.Version) Service {
	return DefaultService{
		version: version,
	}
}

type getVersionRequest struct {
}

type getVersionResponse struct {
	Version v1.Version `json:"version,omitempty"`
	Err     error      `json:"err,omitempty"`
}
