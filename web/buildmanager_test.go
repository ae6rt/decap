package main

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
)

type DeferredBuildsMock struct {
	DeferralServiceBaseMock
	list         []v1.UserBuildEvent
	deferThis    v1.UserBuildEvent
	captureKey   string
	captureEvent v1.UserBuildEvent
	forceError   bool
}

func (t *DeferredBuildsMock) Defer(e v1.UserBuildEvent) error {
	t.captureEvent = e
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return err
}

func (t *DeferredBuildsMock) List() ([]v1.UserBuildEvent, error) {
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return t.list, err
}

func (t *DeferredBuildsMock) Remove(key string) error {
	t.captureKey = key
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return err
}

/*
type BuildManager interface {
	LaunchBuild(v1.UserBuildEvent) error
	CreatePod(*k8sapi.Pod) error
	DeletePod(podName string) error
xx	DeferBuild(v1.UserBuildEvent) error
	LaunchDeferred(ticker <-chan time.Time)
	ClearDeferredBuild(key string) error
xx	DeferredBuilds() ([]v1.UserBuildEvent, error)
	QueueIsOpen() bool
	CloseQueue()
	OpenQueue()
	PodWatcher()
q}
*/

type BuildManagerProjectManagerMock struct {
	ProjectManagerBaseMock
	projects map[string]v1.Project
}

func (t *BuildManagerProjectManagerMock) Get(key string) *v1.Project {
	v := t.projects[key]
	return &v
}

type BuildManagerLockServiceMock struct {
	LockserviceBaseMock
}

type BuildManagerKuberenetesClientMock struct {
	KubernetesClientBaseMock
}

func TestBuildManagerLaunchBuild(t *testing.T) {
	var tests = []struct {
		event    v1.UserBuildEvent
		projects map[string]v1.Project
	}{
		{
			projects: map[string]v1.Project{
				"ae6rt/p1": v1.Project{Team: "ae6rt", ProjectName: "p1"},
			},
			event: v1.UserBuildEvent{Team: "ae6rt", Project: "p1", Ref: "master"},
		},
	}
	for testNumber, test := range tests {
		projectManager := &BuildManagerProjectManagerMock{}
		deferralService := &DeferredBuildsMock{}
		lockService := &BuildManagerLockServiceMock{}
		kubernetesClient := &BuildManagerKuberenetesClientMock{}

		buildManager := DefaultBuildManager{
			deferralService:  deferralService,
			lockService:      lockService,
			projectManager:   projectManager,
			kubernetesClient: kubernetesClient,
			logger:           log.New(ioutil.Discard, "", 0),
		}

		// hack
		getShutdownChan = make(chan string, 1)
		getShutdownChan <- BuildQueueOpen
		// hack

		err := buildManager.LaunchBuild(test.event)
		if err != nil {
			t.Errorf("Test %d: unexpected error: %v\n", testNumber, err)
		}
	}
}

// Test the default build manager public API
func TestBuildManagerGetDeferredBuilds(t *testing.T) {
	var tests = []struct {
		events     []v1.UserBuildEvent
		forceError bool
	}{
		{
			events: []v1.UserBuildEvent{
				v1.UserBuildEvent{},
			},
		},
		{
			forceError: true,
		},
	}

	for testNumber, test := range tests {
		deferralService := &DeferredBuildsMock{list: test.events, forceError: test.forceError}
		buildManager := DefaultBuildManager{deferralService: deferralService}

		got, err := buildManager.DeferredBuilds()

		if test.forceError {
			if err == nil {
				t.Errorf("Test %d: expecting an error\n", testNumber)
			}
			continue
		}

		if err != nil {
			t.Errorf("Test %d: unexpected error: %v\n", testNumber, err)
		}

		if len(got) != len(test.events) {
			t.Errorf("Test %d: expected %d events, got %d\n", testNumber, len(test.events), len(got))
		}
	}
}

// Test the default build manager public API
func TestBuildManagerDeferBuild(t *testing.T) {
	var tests = []struct {
		event      v1.UserBuildEvent
		forceError bool
	}{
		{
			event: v1.UserBuildEvent{ID: "id"},
		},
		{
			forceError: true,
		},
	}

	for testNumber, test := range tests {
		deferralService := &DeferredBuildsMock{deferThis: test.event, forceError: test.forceError}
		buildManager := DefaultBuildManager{deferralService: deferralService}

		err := buildManager.DeferBuild(test.event)

		if test.forceError {
			if err == nil {
				t.Errorf("Test %d: expecting an error\n", testNumber)
			}
			continue
		}

		if err != nil {
			t.Errorf("Test %d: unexpected error: %v\n", testNumber, err)
		}

		if deferralService.captureEvent.ID != test.event.ID {
			t.Errorf("Test %d: want %s, got %s\n", testNumber, test.event.ID, deferralService.captureEvent.ID)
		}
	}
}
