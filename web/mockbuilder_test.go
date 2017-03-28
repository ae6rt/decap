package main

import (
	"time"

	"github.com/ae6rt/decap/web/api/v1"
	k8sapi "k8s.io/client-go/pkg/api/v1"
)

type MockBuilder struct {
	deferred    []v1.UserBuildEvent
	event       v1.UserBuildEvent
	buildID     string
	deferralKey string
	err         error
}

func (d *MockBuilder) LaunchBuild(p v1.UserBuildEvent) error {
	d.event = p
	return nil
}

func (d *MockBuilder) DeletePod(podName string) error {
	d.buildID = podName
	return nil
}

func (d *MockBuilder) DeferBuild(event v1.UserBuildEvent) error {
	return nil
}

func (d *MockBuilder) DeferredBuilds() ([]v1.UserBuildEvent, error) {
	return d.deferred, nil
}

func (d *MockBuilder) CreatePod(pod *k8sapi.Pod) error {
	return nil
}

func (d *MockBuilder) Init() error {
	return nil
}

func (d *MockBuilder) PodWatcher() {
}

func (d *MockBuilder) LaunchDeferred(ticker <-chan time.Time) {
}

func (d *MockBuilder) ClearDeferredBuild(key string) error {
	d.deferralKey = key
	return d.err
}
