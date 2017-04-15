package main

import (
	"time"

	"github.com/ae6rt/decap/web/api/v1"
	k8sapi "k8s.io/client-go/pkg/api/v1"
)

// BuildManager
type BuildManagerBaseMock struct {
}

func (d *BuildManagerBaseMock) LaunchBuild(p v1.UserBuildEvent) error {
	return nil
}

func (d *BuildManagerBaseMock) DeletePod(podName string) error {
	return nil
}

func (d *BuildManagerBaseMock) DeferBuild(event v1.UserBuildEvent) error {
	return nil
}

func (d *BuildManagerBaseMock) DeferredBuilds() ([]v1.UserBuildEvent, error) {
	return nil, nil
}

func (d *BuildManagerBaseMock) CreatePod(pod *k8sapi.Pod) error {
	return nil
}

func (d *BuildManagerBaseMock) PodWatcher() {
}

func (d *BuildManagerBaseMock) LaunchDeferred(ticker <-chan time.Time) {
}

func (d *BuildManagerBaseMock) ClearDeferredBuild(key string) error {
	return nil
}

func (d *BuildManagerBaseMock) QueueIsOpen() bool {
	return true
}

func (d *BuildManagerBaseMock) OpenQueue() {
}

func (d *BuildManagerBaseMock) CloseQueue() {
}

// ProjectManager
type ProjectManagerBaseMock struct {
}

func (t *ProjectManagerBaseMock) Assemble() (map[string]v1.Project, error) {
	return nil, nil
}

func (t *ProjectManagerBaseMock) Set(map[string]v1.Project) {
}

func (t *HooksProjects) Get(key string) *v1.Project {
	return nil
}

func (t *ProjectManagerBaseMock) RepositoryURL() string {
	return ""
}

func (t *ProjectManagerBaseMock) RepositoryBranch() string {
	return ""
}

// DeferralService

type DeferralServiceBaseMock struct {
}

func (t DeferralServiceBaseMock) Defer(event v1.UserBuildEvent) error {
	return nil
}

func (t DeferralServiceBaseMock) Poll() ([]v1.UserBuildEvent, error) {
	return nil, nil
}

func (t DeferralServiceBaseMock) List() ([]v1.UserBuildEvent, error) {
	return nil, nil
}

func (t DeferralServiceBaseMock) Remove(id string) error {
	return nil
}

// LockService

type LockserviceBaseMock struct {
}

func (t LockserviceBaseMock) Acquire(event v1.UserBuildEvent) error {
	return nil
}

func (t LockserviceBaseMock) Release(event v1.UserBuildEvent) error {
	return nil
}
