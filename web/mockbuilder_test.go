package main

import (
	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/locks"
)

type MockBuilder struct {
	deferred []locks.Deferral
	event    BuildEvent
	buildID  string
}

func (d *MockBuilder) LaunchBuild(p BuildEvent) error {
	d.event = p
	return nil
}

func (d *MockBuilder) DeletePod(podName string) error {
	d.buildID = podName
	return nil
}

func (d *MockBuilder) DeferBuild(event BuildEvent, branch string) error {
	return nil
}

func (d *MockBuilder) DeferredBuilds() ([]locks.Deferral, error) {
	return d.deferred, nil
}

func (d *MockBuilder) SquashDeferred([]locks.Deferral) ([]v1.UserBuildEvent, []string) {
	return nil, nil
}

func (d *MockBuilder) ClearDeferredBuild(key string) error {
	return nil
}
