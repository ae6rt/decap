package main

type StorageService interface {
	GetBuildsByProject(project Project, sinceUnixTime uint64, limit uint64) ([]Build, error)
	GetArtifacts(buildID string) ([]byte, error)
	GetConsoleLog(buildID string) ([]byte, error)
}
