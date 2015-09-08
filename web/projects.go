package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ae6rt/gittools"
	"github.com/ae6rt/retry"
)

var projectMutex = &sync.Mutex{}

func findProjects(scriptsRepo, scriptsRepoBranch string) ([]Project, error) {
	projects := make([]Project, 0)
	work := func() error {
		Log.Printf("Finding projects via clone of the build-scripts repository\n")
		cloneDirectory, err := ioutil.TempDir("", "repoclone-")
		defer func() {
			os.RemoveAll(cloneDirectory)
		}()

		if err != nil {
			return err
		}
		if err := gittools.Clone(scriptsRepo, scriptsRepoBranch, cloneDirectory, true); err != nil {
			return err
		}

		buildScripts, err := findBuildScripts(cloneDirectory)
		if err != nil {
			return err
		}

		for _, v := range buildScripts {
			parts := strings.Split(v, "/")
			projects = append(projects, Project{
				Parent:     parts[len(parts)-3],
				Library:    parts[len(parts)-2],
				Descriptor: projectDescriptor(v),
			})
		}
		return nil
	}

	err := retry.New(5*time.Second, 10, retry.DefaultBackoffFunc).Try(work)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func projectDescriptor(scriptPath string) ProjectDescriptor {
	// returning an empty descriptor is acceptable, which is what happens on error
	dpath := descriptorPath(scriptPath)
	var descriptor ProjectDescriptor
	data, err := ioutil.ReadFile(dpath)
	if err != nil {
		return descriptor
	}
	json.Unmarshal(data, &descriptor)
	return descriptor
}

func parentPath(fileName string) string {
	return fileName[:strings.LastIndex(fileName, "/")]
}

func descriptorPath(scriptPath string) string {
	return parentPath(scriptPath) + "/project.json"
}

func getProjects() []Project {
	p := make([]Project, 0)
	projectMutex.Lock()
	p = append(p, projects...)
	projectMutex.Unlock()
	return p
}

func setProjects(p []Project) {
	projectMutex.Lock()
	projects = p
	projectMutex.Unlock()
}
