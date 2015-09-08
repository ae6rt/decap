package main

import (
	"encoding/json"
	"github.com/ae6rt/gittools"
	"github.com/ae6rt/retry"
	"io/ioutil"
	"strings"
	"time"
)

func findProjects(scriptsRepo string) ([]Project, error) {
	projects := make([]Project, 0)
	work := func() error {
		Log.Printf("Finding projects via clone of the build-scripts repository\n")
		cloneDirectory, err := ioutil.TempDir("", "repoclone-")
		if err != nil {
			return err
		}
		if err := gittools.Clone(*buildScriptsRepo, *buildScriptsRepoBranch, cloneDirectory, true); err != nil {
			return err
		}

		buildScripts, err := findBuildScripts(cloneDirectory)
		if err != nil {
			return err
		}

		for _, v := range buildScripts {
			parts := strings.Split(v, "/")
			project := Project{Parent: parts[len(parts)-3], Library: parts[len(parts)-2]}
			parentDir := v[:strings.LastIndex(v, "/")]
			projectDescriptor := parentDir + "/project.json"
			data, err := ioutil.ReadFile(projectDescriptor)
			var descriptor ProjectDescriptor
			if err == nil {
				nerr := json.Unmarshal(data, &descriptor)
				if nerr == nil {
					project.Descriptor = descriptor
				}
			}
			projects = append(projects, project)
		}
		return nil
	}

	err := retry.New(5*time.Second, 10, retry.DefaultBackoffFunc).Try(work)
	if err != nil {
		return nil, err
	}
	return projects, nil
}
