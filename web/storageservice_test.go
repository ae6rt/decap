package main

import "github.com/ae6rt/decap/web/api/v1"

type MockStorageService struct {
	project       v1.Project
	sinceUnixTime uint64
	limit         uint64
	buildID       string
	data          []byte
	err           error
	builds        []v1.Build
}

func (s *MockStorageService) GetBuildsByProject(project v1.Project, sinceTime uint64, limit uint64) ([]v1.Build, error) {
	s.project = project
	s.sinceUnixTime = sinceTime
	s.limit = limit
	return s.builds, s.err
}

func (s *MockStorageService) GetBuildsBuildling() ([]v1.Build, error) {
	return s.builds, s.err
}

func (s *MockStorageService) GetArtifacts(buildID string) ([]byte, error) {
	s.buildID = buildID
	return s.data, s.err
}

func (s *MockStorageService) GetConsoleLog(buildID string) ([]byte, error) {
	s.buildID = buildID
	return s.data, s.err
}
