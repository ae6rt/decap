package main

import "github.com/ae6rt/decap/web/api/v1"

type MockLockService struct {
}

func (t MockLockService) Acquire(event v1.UserBuildEvent) error {
	return nil
}

func (t MockLockService) Release(event v1.UserBuildEvent) error {
	return nil
}

type MockDeferralService struct {
}

func (t MockDeferralService) Defer(event v1.UserBuildEvent) error {
	return nil
}

func (t MockDeferralService) Poll() ([]v1.UserBuildEvent, error) {
	return nil, nil
}

func (t MockDeferralService) List() ([]v1.UserBuildEvent, error) {
	return nil, nil
}

func (t MockDeferralService) Remove(id string) error {
	return nil
}
