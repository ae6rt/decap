package distrlocks

import "time"

var expires time.Duration = 1 * time.Minute

//var expires time.Duration = 6 * time.Hour

// NewDistributedLock creates a new distributed lock based on the project key and branch.
func NewDistributedLock(projectKey, branch string) DistributedLock {
	return DistributedLock{ProjectKey: projectKey, Branch: branch, Expires: time.Now().Add(expires).Unix()}
}

// Key returns a project and branch specific key for use in naming a distributed lock.
func (l DistributedLock) Key() string {
	return l.ProjectKey + "/" + l.Branch
}
