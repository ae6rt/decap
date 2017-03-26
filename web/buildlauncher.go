package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"encoding/json"
	"net/url"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/distrlocks"
	"github.com/ae6rt/decap/web/k8stypes"
	"github.com/ae6rt/decap/web/uuid"
	"github.com/gorilla/websocket"
)

// NewBuilder is the constructor for a new default Builder instance.
func NewBuilder(apiServerURL, username, password, awsKey, awsSecret, awsRegion string, buildScriptsRepo, buildScriptsRepoBranch string,
	distributedLocker distrlocks.DistributedLockService, deferralService deferrals.DeferralService, logger *log.Logger) Builder {

	tlsConfig := tls.Config{}
	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		Log.Printf("Skipping Kubernetes master TLS verify: %v\n", err)
		tlsConfig.InsecureSkipVerify = true
	} else {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
		Log.Println("Kubernetes master secured with TLS")
	}

	apiClient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tlsConfig}}

	data, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")

	return DefaultBuilder{
		MasterURL:              apiServerURL,
		apiToken:               string(data),
		UserName:               username,
		Password:               password,
		LockService:            distributedLocker,
		DeferralService:        deferralService,
		AWSAccessKeyID:         awsKey,
		AWSAccessSecret:        awsSecret,
		AWSRegion:              awsRegion,
		apiClient:              apiClient,
		maxPods:                10,
		buildScriptsRepo:       buildScriptsRepo,
		buildScriptsRepoBranch: buildScriptsRepoBranch,
		tlsConfig:              &tlsConfig,
		logger:                 logger,
	}
}

func (builder DefaultBuilder) makeBaseContainer(buildEvent v1.UserBuildEvent, projects map[string]v1.Project) k8stypes.Container {
	projectKey := buildEvent.Key()
	return k8stypes.Container{
		Name:  "build-server",
		Image: projects[projectKey].Descriptor.Image,
		VolumeMounts: []k8stypes.VolumeMount{
			k8stypes.VolumeMount{
				Name:      "build-scripts",
				MountPath: "/home/decap/buildscripts",
			},
			k8stypes.VolumeMount{
				Name:      "decap-credentials",
				MountPath: "/etc/secrets",
			},
		},
		Env: []k8stypes.EnvVar{
			k8stypes.EnvVar{
				Name:  "BUILD_ID",
				Value: buildEvent.ID,
			},
			k8stypes.EnvVar{
				Name:  "PROJECT_KEY",
				Value: projectKey,
			},
			k8stypes.EnvVar{
				Name:  "BRANCH_TO_BUILD",
				Value: buildEvent.Ref_,
			},
			k8stypes.EnvVar{
				Name:  "BUILD_LOCK_KEY",
				Value: buildEvent.Key(),
			},
			k8stypes.EnvVar{
				Name:  "AWS_ACCESS_KEY_ID",
				Value: builder.AWSAccessKeyID,
			},
			k8stypes.EnvVar{
				Name:  "AWS_SECRET_ACCESS_KEY",
				Value: builder.AWSAccessSecret,
			},
			k8stypes.EnvVar{
				Name:  "AWS_DEFAULT_REGION",
				Value: builder.AWSRegion,
			},
		},
	}
}

func (builder DefaultBuilder) makeSidecarContainers(buildEvent v1.UserBuildEvent, projects map[string]v1.Project) []k8stypes.Container {
	projectKey := buildEvent.Key()
	arr := make([]k8stypes.Container, len(projects[projectKey].Sidecars))

	for i, v := range projects[projectKey].Sidecars {
		var c k8stypes.Container
		err := json.Unmarshal([]byte(v), &c)
		if err != nil {
			Log.Println(err)
			continue
		}
		arr[i] = c
	}
	return arr
}

func (builder DefaultBuilder) makePod(buildEvent v1.UserBuildEvent, buildID, branch string, containers []k8stypes.Container) k8stypes.Pod {
	return k8stypes.Pod{
		TypeMeta: k8stypes.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: k8stypes.ObjectMeta{
			Name:      buildID,
			Namespace: "decap",
			Labels: map[string]string{
				"type":    "decap-build",
				"team":    buildEvent.Team(),
				"project": buildEvent.Project(),
				"branch":  branch,
			},
		},
		Spec: k8stypes.PodSpec{
			Volumes: []k8stypes.Volume{
				k8stypes.Volume{
					Name: "build-scripts",
					VolumeSource: k8stypes.VolumeSource{
						GitRepo: &k8stypes.GitRepoVolumeSource{
							Repository: builder.buildScriptsRepo,
							Revision:   builder.buildScriptsRepoBranch,
						},
					},
				},
				k8stypes.Volume{
					Name: "decap-credentials",
					VolumeSource: k8stypes.VolumeSource{
						Secret: &k8stypes.SecretVolumeSource{
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

func (builder DefaultBuilder) makeContainers(buildEvent v1.UserBuildEvent, projects map[string]v1.Project) []k8stypes.Container {
	baseContainer := builder.makeBaseContainer(buildEvent, projects)
	sidecars := builder.makeSidecarContainers(buildEvent, projects)

	var containers []k8stypes.Container
	containers = append(containers, baseContainer)
	containers = append(containers, sidecars...)
	return containers
}

// LaunchBuild assembles the pod definition, including the base container and sidecars, and calls
// for the pod creation in the cluster.
func (builder DefaultBuilder) LaunchBuild(buildEvent v1.UserBuildEvent) error {

	switch <-getShutdownChan {
	case BuildQueueClose:
		Log.Printf("Build queue closed: %+v\n", buildEvent)
		return nil
	}

	projectKey := buildEvent.Key()
	projects := getProjects()
	project := projects[projectKey]

	if !project.Descriptor.IsRefManaged(buildEvent.Ref()) {
		if <-getLogLevelChan == LogDebug {
			Log.Printf("Ref %s is not managed on project %s.  Not launching a build.\n", buildEvent.Ref(), projectKey)
		}
		return nil
	}

	buildEvent.ID = uuid.Uuid()
	containers := builder.makeContainers(buildEvent, projects)

	pod := builder.makePod(buildEvent, buildEvent.ID, buildEvent.Ref(), containers)

	podBytes, err := json.Marshal(&pod)
	if err != nil {
		return err
	}

	if err := builder.LockService.Acquire(buildEvent); err != nil {
		Log.Printf("Failed to acquire lock for project %s, branch %s: %v\n", projectKey, buildEvent.Ref(), err)
		if err := builder.DeferralService.Defer(buildEvent); err != nil {
			Log.Printf("Failed to defer build: %s/%s\n", projectKey, buildEvent.Ref())
		} else {
			Log.Printf("Deferred build: %s/%s\n", projectKey, buildEvent.Ref())
		}
		return nil
	}

	if <-getLogLevelChan == LogDebug {
		Log.Printf("Acquired lock on build %s for project %s, branch %s\n", buildEvent.ID, projectKey, buildEvent.Ref())
	}

	if err := builder.CreatePod(podBytes); err != nil {
		if err := builder.LockService.Release(buildEvent); err != nil {
			Log.Printf("Failed to release lock on build %s, project %s, branch %s.  No deferral will be attempted.\n", buildEvent.ID, projectKey, buildEvent.Ref())
			return nil
		}
	}

	Log.Printf("Created pod=%s\n", buildEvent.ID)

	return nil
}

// CreatePod creates a pod in the Kubernetes cluster
func (builder DefaultBuilder) CreatePod(pod []byte) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/namespaces/decap/pods", builder.MasterURL), bytes.NewReader(pod))
	if err != nil {
		Log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if builder.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+builder.apiToken)
	} else {
		req.SetBasicAuth(builder.UserName, builder.Password)
	}

	resp, err := builder.apiClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 201 {
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			Log.Printf("Error reading non-201 response body: %v\n", err)
			return err
		} else {
			Log.Printf("%s\n", string(data))
			return nil
		}
	}
	return nil
}

// DeletePod removes the Pod from the Kubernetes cluster
func (builder DefaultBuilder) DeletePod(podName string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/namespaces/decap/pods/%s", builder.MasterURL, podName), nil)
	if err != nil {
		Log.Println(err)
		return err
	}
	if builder.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+builder.apiToken)
	} else {
		req.SetBasicAuth(builder.UserName, builder.Password)
	}

	resp, err := builder.apiClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Log.Printf("Error reading non-200 response body: %v\n", err)
			return err
		}
		Log.Printf("%s\n", string(data))
		return nil
	}
	return nil
}

// Podwatcher watches the k8s master API for pod events.
func (builder DefaultBuilder) PodWatcher() {
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = builder.tlsConfig

	var host string
	{
		u, err := url.Parse(builder.MasterURL)
		if err != nil {
			log.Fatalf("Error parsing master host URL: %s, %s", builder.MasterURL, err)
		}
		host = u.Host
	}

	u, err := url.Parse("wss://" + host + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
	if err != nil {
		log.Fatalf("Error parsing wss:// websocket URL: %s, %s", builder.MasterURL, err)
	}

	var conn *websocket.Conn
	for {
		var resp *http.Response
		var err error

		conn, resp, err = dialer.Dial(u.String(), http.Header{
			"Origin":        []string{"https://" + u.Host},
			"Authorization": []string{"Bearer " + builder.apiToken},
		})

		if err != nil {
			log.Printf("websocket dialer error: %+v: %s", resp, err.Error())
			time.Sleep(5 * time.Second)
		} else {
			defer func() {
				_ = conn.Close()
			}()
			break
		}
	}

	type PodWatch struct {
		Object struct {
			Meta       k8stypes.TypeMeta   `json:",inline"`
			ObjectMeta k8stypes.ObjectMeta `json:"metadata,omitempty"`
			Status     k8stypes.PodStatus  `json:"status"`
		} `json:"object"`
	}

	for {
		_, msg, err := conn.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			continue
		}

		var pod PodWatch
		if err := json.Unmarshal([]byte(msg), &pod); err != nil {
			Log.Println(err)
			continue
		}

		var deletePod bool
		for _, status := range pod.Object.Status.ContainerStatuses {
			if status.Name == "build-server" && status.State.Terminated != nil && status.State.Terminated.ContainerID != "" {
				deletePod = true
				break
			}
		}

		if deletePod {
			if err := builder.DeletePod(pod.Object.ObjectMeta.Name); err != nil {
				Log.Print(err)
			} else {
				Log.Printf("Pod deleted: %s\n", pod.Object.ObjectMeta.Name)
			}
		}
	}
}

// DeferBuild puts the build event on the deferral queue.
func (builder DefaultBuilder) DeferBuild(event v1.UserBuildEvent) error {
	return builder.DeferralService.Defer(event)
}

// DeferredBuilds returns the current queue of deferred builds.  Deferred builds
// are deduped, but preserve the time order of unique entries.
func (builder DefaultBuilder) DeferredBuilds() ([]v1.UserBuildEvent, error) {
	return builder.DeferralService.List()
}

// ClearDeferredBuild removes builds with the given key from the deferral queue.  If more than one
// build in the queue has this key, they will all be removed.
func (builder DefaultBuilder) ClearDeferredBuild(key string) error {
	if err := builder.DeferralService.Remove(key); err != nil {
		return err
	}
	return nil
}

// LaunchDeferred is wrapped in a goroutine, and reads deferred builds from storage and attempts a relaunch of each.
func (builder DefaultBuilder) LaunchDeferred(ticker <-chan time.Time) {
	for _ = range ticker {
		deferredBuilds, err := builder.DeferralService.Poll()
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
