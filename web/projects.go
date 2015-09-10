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

var projects map[string]Project
var projectMutex = &sync.Mutex{}

func findProjects(scriptsRepo, scriptsRepoBranch string) (map[string]Project, error) {
	projects := make(map[string]Project, 0)
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
			parent := parentPath(v)
			sidecars, err := findSidecars(parent)
			if err != nil {
				Log.Println(err)
			}
			var cars string
			for _, sc := range sidecars {
				data, err := ioutil.ReadFile(sc)
				if err != nil {
					Log.Println(err)
				} else {
					cars = cars + "," + string(data)
				}
			}

			p := Project{
				Parent:     parts[len(parts)-3],
				Library:    parts[len(parts)-2],
				Descriptor: projectDescriptor(v),
				Sidecars:   cars,
			}

			key := parts[len(parts)-3] + "/" + parts[len(parts)-2]
			projects[key] = p
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

func getProjects() map[string]Project {
	p := make(map[string]Project, 0)
	projectMutex.Lock()
	for k, v := range projects {
		p[k] = v
	}
	projectMutex.Unlock()
	return p
}

func setProjects(p map[string]Project) {
	projectMutex.Lock()
	projects = p
	projectMutex.Unlock()
}

func findProject(parent, library string) (Project, bool) {
	pr := getProjects()
	p, ok := pr[parent+"/"+library]
	return p, ok
}
