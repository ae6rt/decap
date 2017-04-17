package main

import (
	"fmt"
	"log"
	"time"

	"encoding/json"

	"k8s.io/client-go/pkg/api/unversioned"
	k8sapi "k8s.io/client-go/pkg/api/v1"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/clusterutil"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
	"github.com/ae6rt/decap/web/projects"
	"github.com/ae6rt/decap/web/uuid"
)

// NewBuildManager is the constructor for a new default Builder instance.
func NewBuildManager(
	kubernetesClient KubernetesClient,
	projectManager projects.ProjectManager,
	distributedLocker lock.LockService,
	deferralService deferrals.DeferralService,
	logger *log.Logger,
) BuildManager {
	return DefaultBuildManager{
		lockService:      distributedLocker,
		deferralService:  deferralService,
		kubernetesClient: kubernetesClient,
		projectManager:   projectManager,
		maxPods:          10,
		logger:           logger,
	}
}

// LaunchBuild assembles the pod definition, including the base container and sidecars, and calls
// for the pod creation in the cluster.
func (t DefaultBuildManager) LaunchBuild(buildEvent v1.UserBuildEvent) error {

	if !t.QueueIsOpen() {
		t.logger.Printf("Build queue closed: %+v\n", buildEvent)
		return nil
	}

	projectKey := buildEvent.ProjectKey()

	project := t.projectManager.Get(projectKey)
	if project == nil {
		return fmt.Errorf("Project %s is missing from build scripts repository.\n", projectKey)
	}

	if !project.Descriptor.IsRefManaged(buildEvent.Ref) {
		return fmt.Errorf("Ref %s is not managed on project %s.  Not launching a build.\n", buildEvent.Ref, projectKey)
	}

	buildEvent.ID = uuid.Uuid()

	if err := t.lockService.Acquire(buildEvent); err != nil {
		t.logger.Printf("Failed to acquire lock for project %s, branch %s: %v\n", projectKey, buildEvent.Ref, err)
		if err := t.deferralService.Defer(buildEvent); err != nil {
			t.logger.Printf("Failed to defer build: %s/%s\n", projectKey, buildEvent.Ref)
		} else {
			t.logger.Printf("Deferred build: %s/%s\n", projectKey, buildEvent.Ref)
		}
		return nil
	}

	t.logger.Printf("Acquired lock on build %s for project %s, branch %s\n", buildEvent.ID, projectKey, buildEvent.Ref)

	containers := t.makeContainers(buildEvent)
	pod := t.makePod(buildEvent, containers)

	if err := t.CreatePod(pod); err != nil {
		if err := t.lockService.Release(buildEvent); err != nil {
			t.logger.Printf("Failed to release lock on build %s, project %s, branch %s.  No deferral will be attempted.\n", buildEvent.ID, projectKey, buildEvent.Ref)
			return nil
		}
	}

	t.logger.Printf("Created pod %s\n", buildEvent.ID)

	return nil
}

// CreatePod creates a pod in the Kubernetes cluster
// TODO: this build-job pod will fail to run if the AWS creds are not injected as Secrets.  They had been in env vars.
func (t DefaultBuildManager) CreatePod(pod *k8sapi.Pod) error {
	_, err := t.kubernetesClient.Pods("decap").Create(pod)
	return err
}

// DeletePod removes the Pod from the Kubernetes cluster
func (t DefaultBuildManager) DeletePod(podName string) error {
	err := t.kubernetesClient.Pods("decap").Delete(podName, &k8sapi.DeleteOptions{})
	return err
}

// Podwatcher watches the k8s master API for pod events.
func (t DefaultBuildManager) PodWatcher() {

	Log.Printf("Starting pod watcher")

	deleted := make(map[string]struct{})

	for {
		watched, err := t.kubernetesClient.Pods("decap").Watch(k8sapi.ListOptions{
			LabelSelector: "type=decap-build",
		})
		if err != nil {
			t.logger.Printf("Error watching cluster: %v\n", err)
			continue
		}

		events := watched.ResultChan()

		for event := range events {
			pod := event.Object.(*k8sapi.Pod)

			deletePod := false
			for _, v := range pod.Status.ContainerStatuses {
				if v.Name == "build-server" && v.State.Terminated != nil && v.State.Terminated.ContainerID != "" {
					deletePod = true
					break
				}
			}

			// Try to elete the build pod if it has not already been deleted.
			if _, present := deleted[pod.Name]; !present && deletePod {
				if err := t.kubernetesClient.Pods("decap").Delete(pod.Name, nil); err != nil {
					t.logger.Printf("Error deleting build-server pod: %v\n", err)
				} else {
					t.logger.Printf("Deleted pod %s\n", pod.Name)
				}
				deleted[pod.Name] = struct{}{}
			}
		}
	}
}

// DeferBuild puts the build event on the deferral queue.
func (t DefaultBuildManager) DeferBuild(event v1.UserBuildEvent) error {
	return t.deferralService.Defer(event)
}

// DeferredBuilds returns the current queue of deferred builds.  Deferred builds
// are deduped, but preserve the time order of unique entries.
func (t DefaultBuildManager) DeferredBuilds() ([]v1.UserBuildEvent, error) {
	return t.deferralService.List()
}

// ClearDeferredBuild removes builds with the given key from the deferral queue.  If more than one
// build in the queue has this key, they will all be removed.
func (t DefaultBuildManager) ClearDeferredBuild(key string) error {
	if err := t.deferralService.Remove(key); err != nil {
		return err
	}
	return nil
}

// LaunchDeferred is wrapped in a goroutine, and reads deferred builds from storage and attempts a relaunch of each.
func (t DefaultBuildManager) LaunchDeferred(ticker <-chan time.Time) {
	for _ = range ticker {
		deferredBuilds, err := t.deferralService.Poll()
		if err != nil {
			t.logger.Printf("error retrieving deferred builds: %v\n", err)
		}
		for _, evt := range deferredBuilds {
			err := t.LaunchBuild(evt)
			if err != nil {
				t.logger.Printf("Error launching deferred build: %+v\n", err)
			} else {
				t.logger.Printf("Launched deferred build: %+v\n", evt)
			}
		}
	}
}

func (t DefaultBuildManager) makeBaseContainer(buildEvent v1.UserBuildEvent) k8sapi.Container {
	projectKey := buildEvent.ProjectKey()

	return k8sapi.Container{
		Name:  "build-server",
		Image: t.projectManager.Get(projectKey).Descriptor.Image,
		VolumeMounts: []k8sapi.VolumeMount{
			k8sapi.VolumeMount{
				Name:      "build-scripts",
				MountPath: "/home/decap/buildscripts",
			},
			k8sapi.VolumeMount{
				Name:      "decap-credentials",
				MountPath: "/etc/secrets",
			},
		},
		Env: []k8sapi.EnvVar{
			k8sapi.EnvVar{
				Name:  "BUILD_ID",
				Value: buildEvent.ID,
			},
			k8sapi.EnvVar{
				Name:  "PROJECT_KEY",
				Value: projectKey,
			},
			k8sapi.EnvVar{
				Name:  "BRANCH_TO_BUILD",
				Value: buildEvent.Ref,
			},

			// todo Builds do not manage their own locks now.  Can this be removed?  msp april 2017
			k8sapi.EnvVar{
				Name:  "BUILD_LOCK_KEY",
				Value: buildEvent.Lockname(),
			},
		},
	}
}

func (t DefaultBuildManager) makeSidecarContainers(buildEvent v1.UserBuildEvent) []k8sapi.Container {
	projectKey := buildEvent.ProjectKey()

	sidecars := t.projectManager.Get(projectKey).Sidecars

	arr := make([]k8sapi.Container, len(sidecars))

	for i, v := range sidecars {
		var c k8sapi.Container
		err := json.Unmarshal([]byte(v), &c)
		if err != nil {
			t.logger.Println(err)
			continue
		}
		arr[i] = c
	}
	return arr
}

func (t DefaultBuildManager) makePod(buildEvent v1.UserBuildEvent, containers []k8sapi.Container) *k8sapi.Pod {
	return &k8sapi.Pod{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: k8sapi.ObjectMeta{
			Name:      buildEvent.ID,
			Namespace: "decap",
			Labels: map[string]string{
				"type":     "decap-build",
				"team":     clusterutil.AsLabel(buildEvent.Team),
				"project":  clusterutil.AsLabel(buildEvent.Project),
				"branch":   clusterutil.AsLabel(buildEvent.Ref),
				"lockname": clusterutil.AsLabel(buildEvent.Lockname()),
			},
		},
		Spec: k8sapi.PodSpec{
			Volumes: []k8sapi.Volume{
				k8sapi.Volume{
					Name: "build-scripts",
					VolumeSource: k8sapi.VolumeSource{
						GitRepo: &k8sapi.GitRepoVolumeSource{
							Repository: t.projectManager.RepositoryURL(),
							Revision:   t.projectManager.RepositoryBranch(),
						},
					},
				},
				k8sapi.Volume{
					Name: "decap-credentials",
					VolumeSource: k8sapi.VolumeSource{
						Secret: &k8sapi.SecretVolumeSource{
							SecretName: "decap-credentials",
						},
					},
				},
			},
			Containers:    containers,
			RestartPolicy: "Never",
		},
	}
}

func (t DefaultBuildManager) makeContainers(buildEvent v1.UserBuildEvent) []k8sapi.Container {
	baseContainer := t.makeBaseContainer(buildEvent)
	sidecars := t.makeSidecarContainers(buildEvent)

	var containers []k8sapi.Container
	containers = append(containers, baseContainer)
	containers = append(containers, sidecars...)
	return containers
}

// QueueIsOpen returns true if the build queue is open; false otherwise.
func (t DefaultBuildManager) QueueIsOpen() bool {
	return <-getShutdownChan == "open"
}

// OpenQueue opens the build queue
func (t DefaultBuildManager) OpenQueue() {
	setShutdownChan <- BuildQueueOpen
	t.logger.Println("Build queue is open.")
}

// CloseQueue closes the build queue
func (t DefaultBuildManager) CloseQueue() {
	setShutdownChan <- BuildQueueClose
	t.logger.Println("Build queue is closed.")
}
