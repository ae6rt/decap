package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pborman/uuid"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
)

type StashContainer struct {
	Repository StashRepository  `json:"repository"`
	RefChanges []StashRefChange `json:"refChanges"`

	PushEvent
}

type StashRepository struct {
	Slug    string       `json:"slug"`
	Project StashProject `json:"project"`
}

type StashProject struct {
	Key string `json:"key"`
}

type StashRefChange struct {
	RefID string `json:"refId"`
}

func (stash StashContainer) ProjectKey() string {
	return fmt.Sprintf("%s/%s", stash.Repository.Project, stash.Repository.Slug)
}

func (stash StashContainer) Branches() []string {
	branches := make([]string, 0)
	for _, v := range stash.RefChanges {
		branches = append(branches, strings.ToLower(strings.Replace(v.RefID, "refs/heads/", "", -1)))
	}
	return branches
}

type StashHandler struct {
	Locker Locker
	Handler
}

func (han StashHandler) handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Println(err)
		return
	}
	Log.Printf("post-receive hook received: %s\n", data)

	var stashContainer StashContainer
	if err := json.Unmarshal(data, &stashContainer); err != nil {
		Log.Println(err)
		return
	}
	go han.build(stashContainer)
}

func (han StashHandler) build(pushEvent PushEvent) {
	projectKey := pushEvent.ProjectKey()

	buildPod := BuildPod{
		BuildImage:              *image,
		BuildScriptsGitRepo:     *buildScriptsRepo,
		ProjectKey:              projectKey,
		BuildArtifactBucketName: *buildArtifactBucketName,
		ConsoleLogsBucketName:   *buildConsoleLogsBucketName,
	}

	for _, branch := range pushEvent.Branches() {
		buildPod.BranchToBuild = branch
		buildID := uuid.NewRandom().String()
		buildPod.BuildID = buildID

		tmpl, err := template.New("pod").Parse(podTemplate)
		if err != nil {
			Log.Println(err)
			continue
		}

		hydratedTemplate := bytes.NewBufferString("")
		err = tmpl.Execute(hydratedTemplate, buildPod)
		if err != nil {
			Log.Println(err)
			continue
		}

		lockKey := lockKey(projectKey, branch)

		resp, err := han.Locker.Lock(lockKey, buildID)
		if err != nil {
			Log.Println(err)
			continue
		}

		if resp.Node.Value == buildID {
			Log.Printf("Acquired lock on build %s with key %s\n", buildID, lockKey)
			if podError := createPod(hydratedTemplate.Bytes()); podError != nil {
				Log.Println(podError)
				if _, err := han.Locker.Unlock(lockKey, buildID); err != nil {
					Log.Println(err)
				} else {
					Log.Printf("Released lock on build %s with key %s because of pod creation error %v\n", buildID, lockKey, podError)
				}
			}
			Log.Printf("Created pod=%s\n", buildID)
		}
	}
}
