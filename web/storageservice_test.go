package main

type MockStorageService struct {
	project       Project
	sinceUnixTime uint64
	limit         uint64
	buildID       string
	data          []byte
	err           error
	builds        []Build
}

func (s *MockStorageService) GetBuildsByAtom(project Project, sinceTime uint64, limit uint64) ([]Build, error) {
	s.project = project
	s.sinceUnixTime = sinceTime
	s.limit = limit
	return s.builds, s.err
}

func (s *MockStorageService) GetBuildsBuildling() ([]Build, error) {
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
