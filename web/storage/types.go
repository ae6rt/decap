package storage

import "github.com/ae6rt/decap/web/api/v1"

// StorageService models the interaction between Decap and the persistent storage engine that stores build console logs, artifacts, and specific
// build metadata.
type StorageService interface {
	GetBuildsByProject(project v1.Project, sinceUnixTime uint64, limit uint64) ([]v1.Build, error)
	GetArtifacts(buildID string) ([]byte, error)
	GetConsoleLog(buildID string) ([]byte, error)
}
