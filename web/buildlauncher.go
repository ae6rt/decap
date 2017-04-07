package main

import (
	"io/ioutil"
	"log"
	"strings"
	"time"

	"encoding/json"

	"k8s.io/client-go/pkg/api/unversioned"
	k8sapi "k8s.io/client-go/pkg/api/v1"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
	"github.com/ae6rt/decap/web/uuid"
)

// NewBuildLauncher is the constructor for a new default Builder instance.
func NewBuildLauncher(
	kubernetesClient KubernetesClient,
	buildScripts BuildScripts,
	distributedLocker lock.DistributedLockService,
	deferralService deferrals.DeferralService,
	logger *log.Logger,
) Builder {
	return DefaultBuilder{
		lockService:      distributedLocker,
		deferralService:  deferralService,
		kubernetesClient: kubernetesClient,
		buildScripts:     buildScripts,
		maxPods:          10,
		logger:           logger,
	}
}

// LaunchBuild assembles the pod definition, including the base container and sidecars, and calls
// for the pod creation in the cluster.
func (builder DefaultBuilder) LaunchBuild(buildEvent v1.UserBuildEvent) error {

	switch <-getShutdownChan {
	case BuildQueueClose:
		Log.Printf("Build queue closed: %+v\n", buildEvent)
		return nil
	}

	projectKey := buildEvent.Lockname()
	projects := getProjects()
	project := projects[projectKey]

	if !project.Descriptor.IsRefManaged(buildEvent.Ref) {
		if <-getLogLevelChan == LogDebug {
			Log.Printf("Ref %s is not managed on project %s.  Not launching a build.\n", buildEvent.Ref, projectKey)
		}
		return nil
	}

	buildEvent.ID = uuid.Uuid()

	if err := builder.lockService.Acquire(buildEvent); err != nil {
		Log.Printf("Failed to acquire lock for project %s, branch %s: %v\n", projectKey, buildEvent.Ref, err)
		if err := builder.deferralService.Defer(buildEvent); err != nil {
			Log.Printf("Failed to defer build: %s/%s\n", projectKey, buildEvent.Ref)
		} else {
			Log.Printf("Deferred build: %s/%s\n", projectKey, buildEvent.Ref)
		}
		return nil
	}

	if <-getLogLevelChan == LogDebug {
		Log.Printf("Acquired lock on build %s for project %s, branch %s\n", buildEvent.ID, projectKey, buildEvent.Ref)
	}

	containers := builder.makeContainers(buildEvent, projects)
	pod := builder.makePod(buildEvent, containers)
	if err := builder.CreatePod(pod); err != nil {
		if err := builder.lockService.Release(buildEvent); err != nil {
			Log.Printf("Failed to release lock on build %s, project %s, branch %s.  No deferral will be attempted.\n", buildEvent.ID, projectKey, buildEvent.Ref)
			return nil
		}
	}

	Log.Printf("Created pod=%s\n", buildEvent.ID)

	return nil
}

// CreatePod creates a pod in the Kubernetes cluster
// TODO: this build-job pod will fail to run if the AWS creds are not injected as Secrets.  They had been in env vars.
func (builder DefaultBuilder) CreatePod(pod *k8sapi.Pod) error {
	_, err := builder.kubernetesClient.Pods("decap").Create(pod)
	return err
}

// DeletePod removes the Pod from the Kubernetes cluster
func (builder DefaultBuilder) DeletePod(podName string) error {
	err := builder.kubernetesClient.Pods("decap").Delete(podName, &k8sapi.DeleteOptions{})
	return err
}

// Podwatcher watches the k8s master API for pod events.
func (builder DefaultBuilder) PodWatcher() {
	for {
		watched, err := builder.kubernetesClient.Pods("decap").Watch(k8sapi.ListOptions{
			LabelSelector: "type=decap-build",
		})
		if err != nil {
			Log.Printf("Error watching cluster: %v\n", err)
			continue
		}

		events := watched.ResultChan()
		for event := range events {
			pod := event.Object.(*k8sapi.Pod)
			var deletePod bool
			for _, v := range pod.Status.ContainerStatuses {
				if v.Name == "build-server" && v.State.Terminated != nil && v.State.Terminated.ContainerID != "" {
					deletePod = true
					break
				}
			}
			if deletePod {
				if err := builder.kubernetesClient.Pods("decap").Delete(pod.Name, nil); err != nil {
					Log.Printf("Error deleting build-server pod: %v\n", err)
				}
			}
		}
	}
}

// DeferBuild puts the build event on the deferral queue.
func (builder DefaultBuilder) DeferBuild(event v1.UserBuildEvent) error {
	return builder.deferralService.Defer(event)
}

// DeferredBuilds returns the current queue of deferred builds.  Deferred builds
// are deduped, but preserve the time order of unique entries.
func (builder DefaultBuilder) DeferredBuilds() ([]v1.UserBuildEvent, error) {
	return builder.deferralService.List()
}

// ClearDeferredBuild removes builds with the given key from the deferral queue.  If more than one
// build in the queue has this key, they will all be removed.
func (builder DefaultBuilder) ClearDeferredBuild(key string) error {
	if err := builder.deferralService.Remove(key); err != nil {
		return err
	}
	return nil
}

// LaunchDeferred is wrapped in a goroutine, and reads deferred builds from storage and attempts a relaunch of each.
func (builder DefaultBuilder) LaunchDeferred(ticker <-chan time.Time) {
	for _ = range ticker {
		deferredBuilds, err := builder.deferralService.Poll()
		if err != nil {
			builder.logger.Printf("error retrieving deferred builds: %v\n", err)
		}
		for _, evt := range deferredBuilds {
			err := builder.LaunchBuild(evt)
			if err != nil {
				Log.Printf("Error launching deferred build: %+v\n", err)
			} else {
				Log.Printf("Launched deferred build: %+v\n", evt)
			}
		}
	}
}

func kubeSecret(file string, defaultValue string) string {
	v, err := ioutil.ReadFile(file)
	if err != nil {
		Log.Printf("Secret %s not found in the filesystem.  Using default.\n", file)
		return defaultValue
	}
	Log.Printf("Successfully read secret %s from the filesystem\n", file)
	return string(v)
}

func (builder DefaultBuilder) makeBaseContainer(buildEvent v1.UserBuildEvent, projects map[string]v1.Project) k8sapi.Container {
	projectKey := buildEvent.ProjectKey()
	return k8sapi.Container{
		Name:  "build-server",
		Image: projects[projectKey].Descriptor.Image,
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
			k8sapi.EnvVar{
				Name:  "BUILD_LOCK_KEY",
				Value: buildEvent.Lockname(),
			},
		},
	}
}

func (builder DefaultBuilder) makeSidecarContainers(buildEvent v1.UserBuildEvent, projects map[string]v1.Project) []k8sapi.Container {
	projectKey := buildEvent.ProjectKey()
	arr := make([]k8sapi.Container, len(projects[projectKey].Sidecars))

	for i, v := range projects[projectKey].Sidecars {
		var c k8sapi.Container
		err := json.Unmarshal([]byte(v), &c)
		if err != nil {
			Log.Println(err)
			continue
		}
		arr[i] = c
	}
	return arr
}

func (builder DefaultBuilder) makePod(buildEvent v1.UserBuildEvent, containers []k8sapi.Container) *k8sapi.Pod {
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
				"team":     asLabel(buildEvent.Team),
				"project":  asLabel(buildEvent.Project),
				"branch":   asLabel(buildEvent.Ref),
				"lockname": asLabel(buildEvent.Lockname()),
			},
		},
		Spec: k8sapi.PodSpec{
			Volumes: []k8sapi.Volume{
				k8sapi.Volume{
					Name: "build-scripts",
					VolumeSource: k8sapi.VolumeSource{
						GitRepo: &k8sapi.GitRepoVolumeSource{
							Repository: builder.buildScripts.URL,
							Revision:   builder.buildScripts.Branch,
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

func (builder DefaultBuilder) makeContainers(buildEvent v1.UserBuildEvent, projects map[string]v1.Project) []k8sapi.Container {
	baseContainer := builder.makeBaseContainer(buildEvent, projects)
	sidecars := builder.makeSidecarContainers(buildEvent, projects)

	var containers []k8sapi.Container
	containers = append(containers, baseContainer)
	containers = append(containers, sidecars...)
	return containers
}

func asLabel(s string) string {
	forbidden := []string{".", "-", "/"}
	t := s
	for _, v := range forbidden {
		t = strings.Replace(t, v, "_", -1)
		t = strings.Replace(t, v, "_", -1)
		t = strings.Replace(t, v, "_", -1)
	}
	return t
}
