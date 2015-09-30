package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"

	"github.com/ae6rt/decap/web/k8stypes"
	"github.com/ae6rt/retry"
	"github.com/pborman/uuid"
	"golang.org/x/net/websocket"
)

func NewDefaultDecap(apiServerURL, username, password, awsKey, awsSecret, awsRegion string, locker Locker, buildScriptsRepo, buildScriptsRepoBranch string) DefaultDecap {
	apiClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	data, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")

	return DefaultDecap{
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

func (decap DefaultDecap) makeBaseContainer(buildEvent BuildEvent, buildID, branch string, projects map[string]Project) k8stypes.Container {
	projectKey := buildEvent.ProjectKey()
	lockKey := decap.Locker.Key(projectKey, branch)
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
				Value: decap.AWSAccessKeyID,
			},
			k8stypes.EnvVar{
				Name:  "AWS_SECRET_ACCESS_KEY",
				Value: decap.AWSAccessSecret,
			},
			k8stypes.EnvVar{
				Name:  "AWS_DEFAULT_REGION",
				Value: decap.AWSRegion,
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

func (decap DefaultDecap) makeSidecarContainers(buildEvent BuildEvent, projects map[string]Project) []k8stypes.Container {
	projectKey := buildEvent.ProjectKey()
	arr := make([]k8stypes.Container, len(projects[projectKey].Sidecars))

	var c k8stypes.Container
	for i, v := range projects[projectKey].Sidecars {
		err := json.Unmarshal([]byte(v), &c)
		if err != nil {
			Log.Println(err)
			continue
		}
		arr[i] = c
	}
	return arr
}

func (decap DefaultDecap) makePod(buildEvent BuildEvent, buildID, branch string, containers []k8stypes.Container) k8stypes.Pod {
	pod := k8stypes.Pod{
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
				"library": buildEvent.Library(),
				"branch":  branch,
			},
		},
		Spec: k8stypes.PodSpec{
			Volumes: []k8stypes.Volume{
				k8stypes.Volume{
					Name: "build-scripts",
					VolumeSource: k8stypes.VolumeSource{
						GitRepo: &k8stypes.GitRepoVolumeSource{
							Repository: decap.buildScriptsRepo,
							Revision:   decap.buildScriptsRepoBranch,
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
	return pod
}

func (k8s DefaultDecap) LaunchBuild(buildEvent BuildEvent) error {
	projectKey := buildEvent.ProjectKey()

	projs := getProjects()

	for _, branch := range buildEvent.Refs() {
		key := k8s.Locker.Key(projectKey, branch)
		buildID := uuid.NewRandom().String()

		baseContainer := k8s.makeBaseContainer(buildEvent, buildID, branch, projs)
		sidecars := k8s.makeSidecarContainers(buildEvent, projs)

		containers := make([]k8stypes.Container, 0)
		containers = append(containers, baseContainer)
		containers = append(containers, sidecars...)

		pod := k8s.makePod(buildEvent, buildID, branch, containers)

		podBytes, err := json.Marshal(&pod)
		if err != nil {
			Log.Println(err)
			continue
		}

		fullyQualifiedKey := "/buildlocks/" + key
		resp, err := k8s.Locker.Lock(fullyQualifiedKey, buildID)
		if err != nil {
			Log.Printf("Failed to acquire lock %s on build %s: %v\n", key, buildID, err)
			deferredBuild := UserBuildEvent{
				TeamFld:    buildEvent.Team(),
				LibraryFld: buildEvent.Library(),
				RefsFld:    []string{branch},
			}
			if err := k8s.DeferBuild(deferredBuild); err != nil {
				Log.Printf("Failed to defer build: %+v\n", deferredBuild)
			} else {
				Log.Printf("Deferred build: %+v\n", deferredBuild)
			}
			continue
		}

		if resp.Node.Value == buildID {
			Log.Printf("Acquired lock on build %s with key %s\n", buildID, key)
			if podError := k8s.CreatePod(podBytes); podError != nil {
				Log.Println(podError)
				if _, err := k8s.Locker.Unlock(fullyQualifiedKey, buildID); err != nil {
					Log.Println(err)
				} else {
					Log.Printf("Released lock on build %s with key %s because of pod creation error %v\n", buildID, key, podError)
				}
			}
			Log.Printf("Created pod=%s\n", buildID)
		} else {
			Log.Printf("Failed to acquire lock %s on build %s\n", key, buildID)
		}
	}
	return nil
}

func (base DefaultDecap) CreatePod(pod []byte) error {
	Log.Printf("spec pod:%+v\n", pod)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/namespaces/decap/pods", base.MasterURL), bytes.NewReader(pod))
	if err != nil {
		Log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if base.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+base.apiToken)
	} else {
		req.SetBasicAuth(base.UserName, base.Password)
	}

	resp, err := base.apiClient.Do(req)
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

func (base DefaultDecap) DeletePod(podName string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/namespaces/decap/pods/%s", base.MasterURL, podName), nil)
	if err != nil {
		Log.Println(err)
		return err
	}
	if base.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+base.apiToken)
	} else {
		req.SetBasicAuth(base.UserName, base.Password)
	}

	resp, err := base.apiClient.Do(req)
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

func (decap DefaultDecap) Websock() {

	var conn *websocket.Conn

	work := func() error {
		originURL, err := url.Parse(decap.MasterURL + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
		if err != nil {
			return err
		}
		serviceURL, err := url.Parse("wss://" + originURL.Host + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
		if err != nil {
			return err
		}

		var hdrs http.Header
		if decap.apiToken != "" {
			hdrs = map[string][]string{"Authorization": []string{"Bearer " + decap.apiToken}}
		} else {
			hdrs = map[string][]string{"Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte(decap.UserName+":"+decap.Password))}}
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
		for _, status := range pod.Object.Status.Statuses {
			if status.Name == "build-server" && status.State.Terminated.ContainerID != "" {
				deletePod = true
				break
			}
		}
		if deletePod {
			_, err := decap.Locker.Lock("/pods/"+pod.Object.Meta.Name, "anyvalue")
			if err == nil {
				if err := decap.DeletePod(pod.Object.Meta.Name); err != nil {
					Log.Print(err)
				} else {
					Log.Printf("Pod deleted: %s\n", pod.Object.Meta.Name)
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

func (decap DefaultDecap) DeferBuild(event UserBuildEvent) error {
	return nil
}
