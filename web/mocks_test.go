package main

import (
	"github.com/ae6rt/decap/web/distrlocks"
)

type MockDistributedLocker struct {
}

func (t MockDistributedLocker) Acquire(d distrlocks.DistributedLock) error {
	return nil
}

func (t MockDistributedLocker) Release(d distrlocks.DistributedLock) error {
	return nil
}

type MockDeferralService struct {
}

func (t MockDeferralService) Defer(project, branch, buildID string) error {
	return nil
}

func (t MockDeferralService) Resubmit() {}
