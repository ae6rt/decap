package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/pborman/uuid"
)

type PushEvent interface {
	Parent() string
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
	Parent                    string
	Library                   string
	BranchToBuild             string
	BuildLockKey              string
}

type Handler interface {
	handle(w http.ResponseWriter, r *http.Request)
}

type Decap interface {
	GetProjects(pageStart, pageLimit int) ([]Project, error)
}
type DefaultDecap struct {
	MasterURL string
	UserName  string // not needed when running in the cluster - use apiToken instead
	Password  string // not needed when running in the cluster - use apiToken instead
	Locker    Locker

	apiToken  string
	apiClient *http.Client

	Decap
}

func NewDefaultDecap(apiServerURL, username, password string, locker Locker) DefaultDecap {
	// todo when running in cluster, provide root certificate via /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
	httpClient = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	data, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")

	return DefaultDecap{
		MasterURL: apiServerURL,
		apiToken:  string(data),
		UserName:  username,
		Password:  password,
		Locker:    locker,
		apiClient: httpClient,
	}
}

func (c DefaultDecap) GetProjects(pageStart, pageLimit int) ([]Project, error) {
	return nil, nil
}

func (k8s DefaultDecap) launchBuild(pushEvent PushEvent) error {
	projectKey := pushEvent.ProjectKey()

	buildPod := BuildPod{
		BuildImage:                *image,
		BuildScriptsGitRepo:       *buildScriptsRepo,
		BuildScriptsGitRepoBranch: *buildScriptsRepoBranch,
		ProjectKey:                projectKey,
		Parent:                    pushEvent.Parent(),
		Library:                   pushEvent.Library(),
	}

	tmpl, err := template.New("pod").Parse(podTemplate)
	if err != nil {
		Log.Println(err)
		return err
	}

	for _, branch := range pushEvent.Branches() {
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
			if podError := k8s.createPod(hydratedTemplate.Bytes()); podError != nil {
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

func (base DefaultDecap) createPod(pod []byte) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/namespaces/default/pods", base.MasterURL), bytes.NewReader(pod))
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

	resp, err := httpClient.Do(req)
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
