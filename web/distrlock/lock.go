package distrlocks

import "time"

var expires = 3 * time.Hour

// NewDistributedLock creates a new distributed lock based on the project key and branch.
func NewDistributedLock(projectKey, branch string) DistributedLock {
	return DistributedLock{
		ProjectKey: projectKey,
		Branch:     branch,
		Expires:    time.Now().Add(expires).Unix(),
	}
}

// Key returns a project and branch specific key for use in naming a distributed lock.
func (lock DistributedLock) Key() string {
	return lock.ProjectKey + "/" + lock.Branch
}
