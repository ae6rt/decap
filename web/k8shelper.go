package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"

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
	Statuses []ContainerStatus `json:"containerStatuses"`
}

type ContainerStatus struct {
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

type BuildEvent interface {
	Team() string
	Library() string
	ProjectKey() string
	Branches() []string
}

type BuildPod struct {
	BuildID                   string
	BuildScriptsGitRepo       string
	BuildScriptsGitRepoBranch string
	BuildImage                string
	ProjectKey                string
	Team                      string
	Library                   string
	BranchToBuild             string
	BuildLockKey              string
	SidecarContainers         []string
	AWSAccessKeyID            string
	AWSAccessSecret           string
	AWSRegion                 string
}

func (buildPod BuildPod) FormatSidecars(sidecars []string) string {
	var s string
	for _, v := range sidecars {
		s = s + "," + v
	}
	return s
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

	buildPod := BuildPod{
		BuildImage:                projs[projectKey].Descriptor.Image,
		BuildScriptsGitRepo:       *buildScriptsRepo,
		BuildScriptsGitRepoBranch: *buildScriptsRepoBranch,
		ProjectKey:                projectKey,
		Team:                      buildEvent.Team(),
		Library:                   buildEvent.Library(),
		SidecarContainers:         projs[projectKey].Sidecars,
		AWSAccessKeyID:            k8s.AWSAccessKeyID,
		AWSAccessSecret:           k8s.AWSAccessSecret,
		AWSRegion:                 k8s.AWSRegion,
	}

	tmpl, err := template.New("pod").Parse(podTemplate)
	if err != nil {
		Log.Println(err)
		return err
	}

	for _, branch := range buildEvent.Branches() {
		key := k8s.Locker.Key(projectKey, branch)

		buildPod.BranchToBuild = branch
		buildPod.BuildID = uuid.NewRandom().String()
		buildPod.BuildLockKey = key

		hydratedTemplate := bytes.NewBufferString("")
		err = tmpl.Execute(hydratedTemplate, buildPod)
		if err != nil {
			Log.Println(err)
			continue
		}

		resp, err := k8s.Locker.Lock(key, buildPod.BuildID)
		if err != nil {
			Log.Printf("Failed to acquire lock %s on build %s: %v\n", key, buildPod.BuildID, err)
			continue
		}

		if resp.Node.Value == buildPod.BuildID {
			Log.Printf("Acquired lock on build %s with key %s\n", buildPod.BuildID, key)
			if podError := k8s.CreatePod(hydratedTemplate.Bytes()); podError != nil {
				Log.Println(podError)
				if _, err := k8s.Locker.Unlock(key, buildPod.BuildID); err != nil {
					Log.Println(err)
				} else {
					Log.Printf("Released lock on build %s with key %s because of pod creation error %v\n", buildPod.BuildID, key, podError)
				}
			}
			Log.Printf("Created pod=%s\n", buildPod.BuildID)
		} else {
			Log.Printf("Failed to acquire lock %s on build %s\n", key, buildPod.BuildID)
		}
	}
	return nil
}

func (base DefaultDecap) CreatePod(pod []byte) error {
	Log.Printf("spec pod:\n%s\n", string(pod))

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
	originURL, err := url.Parse(decap.MasterURL + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
	if err != nil {
		Log.Printf("Error parsing websocket origin URL.  Will be unable to reap exited pods.: %v\n", err)
		return
	}
	serviceURL, err := url.Parse("wss://" + originURL.Host + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
	if err != nil {
		Log.Printf("Error parsing websocket service URL.  Will be unable to reap exited pods.: %v\n", err)
		return
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

	conn, err := websocket.DialConfig(&cfg)
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
			_, err := decap.Locker.Lock(pod.Object.Meta.Name, "anyvalue")
			if err == nil {
				if err := decap.DeletePod(pod.Object.Meta.Name); err != nil {
					Log.Print(err)
				} else {
					Log.Printf("Pod deleted: %s\n", pod.Object.Meta.Name)
				}
			} else {
				Log.Println(err)
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
