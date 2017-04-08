package lock

import "github.com/ae6rt/decap/web/api/v1"

// LockService defines the exported interface that a distributed lock service supports.
type LockService interface {
	Acquire(v1.UserBuildEvent) error
	Release(v1.UserBuildEvent) error
}
