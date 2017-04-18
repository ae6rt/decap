package projects

import "github.com/ae6rt/decap/web/api/v1"

// ProjectManager is the interface to the build scripts repository.
type ProjectManager interface {
	Assemble() error
	GetProjects() map[string]v1.Project
	GetProjectByTeamName(team, project string) (v1.Project, bool)
	Get(key string) *v1.Project
	RepositoryURL() string
	RepositoryBranch() string
}
