package main

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	k8scorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8sapi "k8s.io/client-go/pkg/api/v1"
)

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
}
*/

type LaunchBuildDeferralService struct {
	DeferralServiceBaseMock
	captureDeferEvent v1.UserBuildEvent
	forceErrors       deferralServiceErrors
}

func (t *LaunchBuildDeferralService) Defer(e v1.UserBuildEvent) error {
	t.captureDeferEvent = e
	var err error
	if t.forceErrors.deferral {
		err = errors.New("forced error")
	}
	return err
}

type LaunchBuildProjectManager struct {
	ProjectManagerBaseMock
	projects map[string]v1.Project
}

func (t *LaunchBuildProjectManager) Get(key string) *v1.Project {
	v := t.projects[key]
	return &v
}

type LaunchBuildLockService struct {
	LockserviceBaseMock
	captureEvent v1.UserBuildEvent
	forceErrors  lockServiceErrors
}

func (t *LaunchBuildLockService) Acquire(e v1.UserBuildEvent) error {
	t.captureEvent = e
	var err error
	if t.forceErrors.acquire {
		err = errors.New("forced error")
	}
	return err
}

type LaunchBuildKubernetesClient struct {
	KubernetesClientBaseMock
	podsGetter *LaunchBuildPodsGetter
}

func (t *LaunchBuildKubernetesClient) Pods(ns string) k8scorev1.PodInterface {
	return t.podsGetter
}

type LaunchBuildPodsGetter struct {
	podOps
	capturePod *k8sapi.Pod
	forceError bool
}

func (t *LaunchBuildPodsGetter) Create(pod *k8sapi.Pod) (*k8sapi.Pod, error) {
	t.capturePod = pod
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return nil, err
}

type lockServiceErrors struct {
	acquire bool
	release bool
}

type deferralServiceErrors struct {
	deferral bool
}

type collaboratorErrors struct {
	lockservice lockServiceErrors
	deferralServiceErrors
}

func TestBuildManagerLaunchBuild(t *testing.T) {
	var tests = []struct {
		event              v1.UserBuildEvent
		projects           map[string]v1.Project
		collaboratorErrors collaboratorErrors
	}{
		{
			projects: map[string]v1.Project{
				"ae6rt/p1": v1.Project{Team: "ae6rt", ProjectName: "p1"},
			},
			event:              v1.UserBuildEvent{Team: "ae6rt", Project: "p1", Ref: "master"},
			collaboratorErrors: collaboratorErrors{},
		},
	}

	for testNumber, test := range tests {
		deferralService := &LaunchBuildDeferralService{}
		lockService := &LaunchBuildLockService{}
		projectManager := &LaunchBuildProjectManager{}
		kubernetesClient := &LaunchBuildKubernetesClient{podsGetter: &LaunchBuildPodsGetter{}}

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

		if lockService.captureEvent.ProjectKey() != test.event.ProjectKey() {
			t.Errorf("Test %d: want %s, got %s\n", testNumber, test.event.ProjectKey(), lockService.captureEvent.ProjectKey())
		}

		if kubernetesClient.podsGetter.capturePod == nil {
			t.Errorf("Test %d: expecting non-nil pod.\n", testNumber)
		}
	}
}

type GetDeferredBuildsDeferralService struct {
	DeferralServiceBaseMock
	list         []v1.UserBuildEvent
	deferThis    v1.UserBuildEvent
	captureKey   string
	captureEvent v1.UserBuildEvent
	forceError   bool
}

func (t *GetDeferredBuildsDeferralService) Defer(e v1.UserBuildEvent) error {
	t.captureEvent = e
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return err
}

func (t *GetDeferredBuildsDeferralService) List() ([]v1.UserBuildEvent, error) {
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return t.list, err
}

func (t *GetDeferredBuildsDeferralService) Remove(key string) error {
	t.captureKey = key
	var err error
	if t.forceError {
		err = errors.New("forced error")
	}
	return err
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
		deferralService := &GetDeferredBuildsDeferralService{list: test.events, forceError: test.forceError}
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
		deferralService := &GetDeferredBuildsDeferralService{deferThis: test.event, forceError: test.forceError}
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
