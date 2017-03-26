package distrlocks

import "github.com/ae6rt/decap/web/api/v1"

// DefaultLockService queries the k8s master to find out if a pod is building a project+branch.
type DefaultLockService struct {
	k8sClient interface{}
}

// Acquire attempts to acquire a lock on the given object
func (t *DefaultLockService) Acquire(obj v1.UserBuildEvent) error {
	return nil
}

// Release is a no op for the default lock service.
func (t *DefaultLockService) Release(obj v1.UserBuildEvent) error {
	return nil
}
