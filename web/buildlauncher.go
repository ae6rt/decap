package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/k8stypes"
	"github.com/ae6rt/decap/web/locks"
	"github.com/ae6rt/retry"
	"github.com/pborman/uuid"
	"golang.org/x/net/websocket"
)

// NewBuilder is the constructor for a new default Builder instance.
func NewBuilder(apiServerURL, username, password, awsKey, awsSecret, awsRegion string, locker locks.Locker, buildScriptsRepo, buildScriptsRepoBranch string) DefaultBuilder {

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

	apiClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tlsConfig,
	}}

	data, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")

	return DefaultBuilder{
		MasterURL:              apiServerURL,
		apiToken:               string(data),
		UserName:               username,
		Password:               password,
		Locker:                 locker,
		AWSAccessKeyID:         awsKey,
		AWSAccessSecret:        awsSecret,
		AWSRegion:              awsRegion,
		apiClient:              apiClient,
		maxPods:                10,
		buildScriptsRepo:       buildScriptsRepo,
		buildScriptsRepoBranch: buildScriptsRepoBranch,
	}
}

func (builder DefaultBuilder) makeBaseContainer(buildEvent BuildEvent, buildID, branch string, projects map[string]v1.Project) k8stypes.Container {
	projectKey := buildEvent.Key()
	lockKey := builder.Locker.Key(projectKey, branch)
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
				Value: buildID,
			},
			k8stypes.EnvVar{
				Name:  "PROJECT_KEY",
				Value: projectKey,
			},
			k8stypes.EnvVar{
				Name:  "BRANCH_TO_BUILD",
				Value: branch,
			},
			k8stypes.EnvVar{
				Name:  "BUILD_LOCK_KEY",
				Value: lockKey,
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
		Lifecycle: &k8stypes.Lifecycle{
			PreStop: &k8stypes.Handler{
				Exec: &k8stypes.ExecAction{
					Command: []string{
						"bctool", "unlock",
						"--lockservice-base-url", "http://lockservice.decap-system:2379",
						"--build-id", buildID,
						"--build-lock-key", lockKey,
					},
				},
			},
		},
	}
}

func (builder DefaultBuilder) makeSidecarContainers(buildEvent BuildEvent, projects map[string]v1.Project) []k8stypes.Container {
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

func (builder DefaultBuilder) makePod(buildEvent BuildEvent, buildID, branch string, containers []k8stypes.Container) k8stypes.Pod {
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

func (builder DefaultBuilder) makeContainers(buildEvent BuildEvent, buildID, branch string, projects map[string]v1.Project) []k8stypes.Container {
	baseContainer := builder.makeBaseContainer(buildEvent, buildID, branch, projects)
	sidecars := builder.makeSidecarContainers(buildEvent, projects)

	var containers []k8stypes.Container
	containers = append(containers, baseContainer)
	containers = append(containers, sidecars...)
	return containers
}

// Attempt to lock a build.  If that fails, defer it.
func (builder DefaultBuilder) lockOrDefer(buildEvent BuildEvent, ref, buildID, key string) (bool, error) {
	if _, err := builder.Locker.Lock(key, buildID); err != nil {
		Log.Printf("Failed to acquire lock %s on build %s (%+v): %v\n", key, buildID, buildEvent, err)
		if err = builder.DeferBuild(buildEvent, ref); err != nil {
			Log.Printf("Failed to defer build %s after failing to acquire a lock on it: %+v\n", buildID, err)
		} else {
			Log.Printf("Deferred build: %+v\n", buildID)
		}
		return false, err
	}
	return true, nil
}

// Attempt to create a build pod on the cluster.  If that fails, clear the lock and defer it.  If it succeeds, clear
// any deferrals.
func (builder DefaultBuilder) createOrDefer(data []byte, buildEvent BuildEvent, buildID, ref, key string) (bool, error) {
	if podError := builder.CreatePod(data); podError != nil {
		Log.Printf("Failed creating pod: %v\n", podError)
		if _, err := builder.Locker.Unlock(key, buildID); err != nil {
			Log.Printf("Failed unlocking build %s after pod creation failed: %v\n", buildID, err)
			if err = builder.DeferBuild(buildEvent, ref); err != nil {
				Log.Printf("Failed deferring build %+v for ref %s after failed unlocking after pod creation attempt: %+v\n", buildEvent, ref, err)
			}
			return false, err
		}
		Log.Printf("Released lock on build %s with key %s because of pod creation error %v\n", buildID, key, podError)
		return false, podError
	}
	return true, nil
}

// LaunchBuild assembles the pod definition, including the base container and sidecars, and calls
// for the pod creation in the cluster.
func (builder DefaultBuilder) LaunchBuild(buildEvent BuildEvent) error {

	switch <-getShutdownChan {
	case BUILD_QUEUE_CLOSE:
		Log.Printf("Build queue closed: %+v\n", buildEvent)
		return nil
	}

	projectKey := buildEvent.Key()
	projects := getProjects()
	project := projects[projectKey]

	for _, ref := range buildEvent.Refs() {
		if !project.Descriptor.IsRefManaged(ref) {
			if <-getLogLevelChan == LOG_DEBUG {
				Log.Printf("Ref %s is not managed on project %s.  Not launching a build.\n", ref, projectKey)
			}
			continue
		}

		key := builder.Locker.Key(projectKey, ref)
		buildID := uuid.NewRandom().String()

		containers := builder.makeContainers(buildEvent, buildID, ref, projects)

		pod := builder.makePod(buildEvent, buildID, ref, containers)

		podBytes, err := json.Marshal(&pod)
		if err != nil {
			Log.Println(err)
			continue
		}

		locked, err := builder.lockOrDefer(buildEvent, ref, buildID, key)
		if err != nil {
			Log.Println(err)
			continue
		}

		if !locked {
			continue
		}

		if <-getLogLevelChan == LOG_DEBUG {
			Log.Printf("Acquired lock on build %s with key %s\n", buildID, key)
		}

		created, err := builder.createOrDefer(podBytes, buildEvent, buildID, ref, key)
		if err != nil {
			Log.Println(err)
			continue
		}

		if created {
			Log.Printf("Created pod=%s\n", buildID)
		}
	}
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
		resp.Body.Close()
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
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			Log.Printf("Error reading non-200 response body: %v\n", err)
			return err
		} else {
			Log.Printf("%s\n", string(data))
			return nil
		}
	}
	return nil
}

func (builder DefaultBuilder) Websock() {

	var conn *websocket.Conn

	type PodWatch struct {
		Object struct {
			Meta       k8stypes.TypeMeta   `json:",inline"`
			ObjectMeta k8stypes.ObjectMeta `json:"metadata,omitempty"`
			Status     k8stypes.PodStatus  `json:"status"`
		} `json:"object"`
	}

	work := func() error {
		originURL, err := url.Parse(builder.MasterURL + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
		if err != nil {
			return err
		}
		serviceURL, err := url.Parse("wss://" + originURL.Host + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
		if err != nil {
			return err
		}

		var hdrs http.Header
		if builder.apiToken != "" {
			hdrs = map[string][]string{"Authorization": []string{"Bearer " + builder.apiToken}}
		} else {
			hdrs = map[string][]string{"Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte(builder.UserName+":"+builder.Password))}}
		}

		cfg := websocket.Config{
			Location:  serviceURL,
			Origin:    originURL,
			TlsConfig: &tls.Config{InsecureSkipVerify: true},
			Header:    hdrs,
			Version:   websocket.ProtocolVersionHybi13,
		}

		if conn, err = websocket.DialConfig(&cfg); err != nil {
			return err
		}
		return nil
	}

	err := retry.New(5*time.Second, 60, retry.DefaultBackoffFunc).Try(work)
	if err != nil {
		Log.Printf("Error opening websocket connection.  Will be unable to reap exited pods.: %v\n", err)
		return
	}
	Log.Print("Watching pods on websocket")

	var msg string
	for {
		err := websocket.Message.Receive(conn, &msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			Log.Println("Couldn't receive msg " + err.Error())
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
			// Mark the pod as deleted in etcd so subsequent events don't drive a 2nd deletion attempt
			_, err := builder.Locker.Lock("/pods/"+pod.Object.ObjectMeta.Name, "anyvalue")

			if err == nil {
				// for now just report on what we would have done vs doing it
				Log.Printf("Would have deleted pod: %s\n", pod.Object.ObjectMeta.Name)
				if true {
					continue
				}

				if err := builder.DeletePod(pod.Object.ObjectMeta.Name); err != nil {
					Log.Print(err)
				} else {
					Log.Printf("Pod deleted: %s\n", pod.Object.ObjectMeta.Name)
				}
			}
		}
	}
}

func (builder DefaultBuilder) DeferBuild(event BuildEvent, branch string) error {
	ube := v1.UserBuildEvent{
		Team_:    event.Team(),
		Project_: event.Project(),
		Refs_:    []string{branch},
	}
	data, _ := json.Marshal(&ube)
	_, err := builder.Locker.Defer(data)
	return err
}

// SquashDeferred takes a in-created-order list of deferred builds and filters out duplicate
// team + project + branch deferrals, returning the first in the list of each unique build event.
func (builder DefaultBuilder) SquashDeferred(deferrals []locks.Deferral) ([]v1.UserBuildEvent, []string) {

	events := make([]v1.UserBuildEvent, len(deferrals))
	for i, deferral := range deferrals {
		var ube v1.UserBuildEvent
		if err := json.Unmarshal([]byte(deferral.Data), &ube); err != nil {
			Log.Printf("Error deserializing build event %s: %v\n", deferral.Data, err)
			continue
		}
		ube.Deferral.Key = deferral.Key
		events[i] = ube
	}

	// h{n} are hashes, c{n} are in-created-order keys
	// h1:c1  < keep
	// h2:c2  < keep
	// h1:c3  < omit
	// h2:c4  < omit
	// h3:c5  < keep

	// find the event hashes
	hashes := make(map[string]string)
	for _, v := range events {
		hashes[v.Hash()] = ""
	}

	// record the position of the first occurrence of a hash in the time-ordered events
	positions := make(map[string]int)
	for k, _ := range hashes {
		for i, j := range events {
			if k == j.Hash() {
				positions[k] = i
				break
			}
		}
	}

	// extract the hash positions and sort ascending to preserve time ordering
	slots := make([]int, len(positions))
	i := 0
	for _, v := range positions {
		slots[i] = v
		i += 1
	}
	sort.Ints(slots)

	// extract from events the object at each slot
	squashed := make([]v1.UserBuildEvent, len(slots))
	for i, j := range slots {
		squashed[i] = events[j]
	}

	// record the deferral key for the omitted events so they can be deleted
	var excluded []string
	for i, k := range deferrals {
		foundIt := false
		for _, j := range slots {
			if i == j {
				foundIt = true
				break
			}
		}
		if !foundIt {
			excluded = append(excluded, k.Key)
		}
	}
	return squashed, excluded
}

func (builder DefaultBuilder) DeferredBuilds() ([]locks.Deferral, error) {
	return builder.Locker.DeferredBuilds()
}

func (builder DefaultBuilder) ClearDeferredBuild(key string) error {
	_, err := builder.Locker.ClearDeferred(key)
	return err
}

// LaunchDeferred is wrapped in a goroutine, and reads deferred builds from storage and attempts a relaunch of each.
func (builder DefaultBuilder) LaunchDeferred(ticker <-chan time.Time) {
	for _ = range ticker {
		if builds, err := builder.DeferredBuilds(); err != nil {
			Log.Println(err)
		} else {
			squashed, excluded := builder.SquashDeferred(builds)
			for _, v := range excluded {
				if err := builder.ClearDeferredBuild(v); err != nil {
					Log.Printf("Failed to clear deferred build for omitted event %+v: %+v\n", v, err)
				}
			}
			for _, build := range squashed {
				var err error
				err = builder.ClearDeferredBuild(build.Deferral.Key)
				if err != nil {
					Log.Printf("Failed to clear deferred build, will not launch: %+v: %v\n", build, err)
					continue
				}
				Log.Printf("Cleared deferred build at key %s\n", build.Deferral.Key)

				err = builder.LaunchBuild(build)
				if err != nil {
					Log.Printf("Error launching deferred build: %+v\n", err)
				} else {
					Log.Printf("Launched deferred build: %+v\n", build)
				}
			}
		}
	}
}

func kubeSecret(file string, defaultValue string) string {
	if v, err := ioutil.ReadFile(file); err != nil {
		Log.Printf("Secret %s not found in the filesystem.  Using default.\n", file)
		return defaultValue
	} else {
		Log.Printf("Successfully read secret %s from the filesystem\n", file)
		return string(v)
	}
}
