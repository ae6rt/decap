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

type PodWatch struct {
	Object Object `json:"object"`
}

type Object struct {
	Meta   Metadata `json:"metadata"`
	Status Status   `json:"status"`
}

type Metadata struct {
	Name string `json:"name"`
}

type Status struct {
	Statuses []XContainerStatus `json:"containerStatuses"`
}

type XContainerStatus struct {
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
	State State  `json:"state"`
}

type State struct {
	Terminated Terminated `json:"terminated"`
}

type Terminated struct {
	ContainerID string `json:"containerID"`
	ExitCode    int    `json:"exitCode"`
}

// TODO distinguish between pushes and branch creation.  Github has a header value that allows these to be differentiated.
// https://developer.github.com/webhooks/#delivery-headers
// https://gist.githubusercontent.com/ae6rt/53a25e726ac00b4cb535/raw/e3f412f6e7f408a56d0d691a1ec8b7658a495124/gh-create.json
// https://gist.githubusercontent.com/ae6rt/2be93f7d5edef8030b52/raw/29f591eb8ecc5555c55f1878b545613c1f9839b7/gh-push.json
type BuildEvent interface {
	Team() string
	Library() string
	ProjectKey() string
	Branches() []string
}

type DefaultDecap struct {
	MasterURL       string
	UserName        string
	Password        string
	AWSAccessKeyID  string
	AWSAccessSecret string
	AWSRegion       string
	Locker          Locker

	apiToken  string
	apiClient *http.Client
}

type RepoManagerCredential struct {
	User     string
	Password string
}

type StorageService interface {
	GetBuildsByProject(project Project, sinceUnixTime uint64, limit uint64) ([]Build, error)
	GetArtifacts(buildID string) ([]byte, error)
	GetConsoleLog(buildID string) ([]byte, error)
}

type Decap interface {
	LaunchBuild(buildEvent BuildEvent) error
	DeletePod(podName string) error
}

func NewDefaultDecap(apiServerURL, username, password, awsKey, awsSecret, awsRegion string, locker Locker) DefaultDecap {
	// todo when running in cluster, provide root certificate via /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
	apiClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	data, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")

	return DefaultDecap{
		MasterURL:       apiServerURL,
		apiToken:        string(data),
		UserName:        username,
		Password:        password,
		Locker:          locker,
		AWSAccessKeyID:  awsKey,
		AWSAccessSecret: awsSecret,
		AWSRegion:       awsRegion,
		apiClient:       apiClient,
	}
}

func (k8s DefaultDecap) LaunchBuild(buildEvent BuildEvent) error {
	projectKey := buildEvent.ProjectKey()

	projs := getProjects()

	for _, branch := range buildEvent.Branches() {
		key := k8s.Locker.Key(projectKey, branch)
		buildID := uuid.NewRandom().String()

		containers := make([]k8stypes.Container, 1+len(projs[projectKey].Sidecars))

		baseContainer := k8stypes.Container{
			Name:  "build-server",
			Image: projs[projectKey].Descriptor.Image,
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
					Value: key,
				},
				k8stypes.EnvVar{
					Name:  "AWS_ACCESS_KEY_ID",
					Value: k8s.AWSAccessKeyID,
				},
				k8stypes.EnvVar{
					Name:  "AWS_SECRET_ACCESS_KEY",
					Value: k8s.AWSAccessSecret,
				},
				k8stypes.EnvVar{
					Name:  "AWS_DEFAULT_REGION",
					Value: k8s.AWSRegion,
				},
			},
			Lifecycle: &k8stypes.Lifecycle{
				PreStop: &k8stypes.Handler{
					Exec: &k8stypes.ExecAction{
						Command: []string{
							"bctool", "unlock",
							"--lockservice-base-url", "http://lockservice.decap-system:2379",
							"--build-id", buildID,
							"--build-lock-key", key,
						},
					},
				},
			},
		}

		containers[0] = baseContainer
		for i, v := range projs[projectKey].Sidecars {
			var c k8stypes.Container
			err := json.Unmarshal([]byte(v), &c)
			if err != nil {
				Log.Println(err)
				continue
			}
			containers[i+1] = c
		}

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
								Repository: *buildScriptsRepo,
								Revision:   *buildScriptsRepoBranch,
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

		podBytes, err := json.Marshal(&pod)
		if err != nil {
			Log.Println(err)
			continue
		}

		fullyQualifiedKey := "/buildlocks/" + key
		resp, err := k8s.Locker.Lock(fullyQualifiedKey, buildID)
		if err != nil {
			Log.Printf("Failed to acquire lock %s on build %s: %v\n", key, buildID, err)
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
