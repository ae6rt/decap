package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/pborman/uuid"
	"html/template"
	"io/ioutil"
	"net/http"
)

type PushEvent interface {
	ProjectKey() string
	Branches() []string
}

type BuildPod struct {
	BuildID                   string
	BuildScriptsGitRepo       string
	BuildScriptsGitRepoBranch string
	BuildImage                string
	ProjectKey                string
	BranchToBuild             string
	BuildLockKey              string
}

type Handler interface {
	handle(w http.ResponseWriter, r *http.Request)
}

type K8sBase struct {
	MasterURL string
	UserName  string
	Password  string
	Locker    Locker

	apiToken  string
	apiClient *http.Client
}

func NewK8s(apiServerURL, username, password string, locker Locker) K8sBase {
	httpClient = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		Log.Printf("No service account token: %v.  Falling back to api server username/password for master authentication.\n", err)
	}

	return K8sBase{
		MasterURL: apiServerURL,
		apiToken:  string(data),
		UserName:  username,
		Password:  password,
		Locker:    locker,
		apiClient: httpClient,
	}
}

func (k8s K8sBase) launchBuild(pushEvent PushEvent) error {
	projectKey := pushEvent.ProjectKey()

	buildPod := BuildPod{
		BuildImage:                *image,
		BuildScriptsGitRepo:       *buildScriptsRepo,
		BuildScriptsGitRepoBranch: *buildScriptsRepoBranch,
		ProjectKey:                projectKey,
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

func (base K8sBase) createPod(pod []byte) error {
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
