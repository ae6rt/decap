package v1

import "fmt"

// Lockname returns the build locknae, which is the ProjectKey concatenated with "/" followed
// by the branch.
func (e UserBuildEvent) Lockname() string {
	return fmt.Sprintf("%s/%s/%s", e.Team_, e.Project_, e.Ref_)
}

// ProjectKey returns a key that identifies the codebase.  For github projects, this is the github account owner
// concatenated with "/" followed by the basename of the repository.
func (e UserBuildEvent) ProjectKey() string {
	return fmt.Sprintf("%s/%s", e.Team_, e.Project_)
}
