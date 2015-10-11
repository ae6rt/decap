package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ae6rt/gittools"
	"github.com/ae6rt/retry"
)

const buildScriptRegex = `build\.sh`
const projectDescriptorRegex = `project\.json`
const sideCarRegex = `^.+-sidecar\.json`

var projectSetChan = make(chan map[string]Project)
var projectGetChan = make(chan map[string]Project)

func projectMux(initialValue map[string]Project) {
	t := initialValue
	Log.Print("Project channel mux running")
	for {
		select {
		case t = <-projectSetChan:
		case projectGetChan <- t:
		}
	}
}

func filesByRegex(root, expression string) ([]string, error) {
	if !strings.HasPrefix(root, "/") {
		return nil, fmt.Errorf("Root must be an absolute path: %s\n", root)
	}
	if strings.HasSuffix(root, "/") {
		return nil, fmt.Errorf("Root must not end in /: %s\n", root)
	}

	regex, err := regexp.Compile(expression)
	if err != nil {
		return nil, err
	}

	// files of interest reside at a fixed depth of 3 below root
	slashOffset := strings.Count(root, "/") + 3

	files := make([]string, 0)
	markFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(path, ".git") {
			return filepath.SkipDir
		}

		// record this as a too-deep path we never want to traverse again
		if info.IsDir() && strings.Count(path, "/") > slashOffset {
			return filepath.SkipDir
		}

		if strings.Count(path, "/") == slashOffset && info.Mode().IsRegular() && regex.MatchString(info.Name()) {
			files = append(files, path)
		}
		return nil
	}

	// Walk relative to root.
	err = filepath.Walk(root, markFn)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func assembleProjects(scriptsRepo, scriptsRepoBranch string) (map[string]Project, error) {
	projects := make(map[string]Project, 0)
	work := func() error {
		Log.Printf("Clone build-scripts repository...\n")
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

		// Build scripts are the anchors for a project.  If build.sh does not exist the project is skipped.
		buildScripts, err := filesByRegex(cloneDirectory, buildScriptRegex)
		if err != nil {
			return err
		}
		descriptorFiles, err := filesByRegex(cloneDirectory, projectDescriptorRegex)
		if err != nil {
			return err
		}
		sidecarFiles, err := filesByRegex(cloneDirectory, sideCarRegex)
		if err != nil {
			return err
		}

		buildScriptMap := indexFilesByTeamProject(buildScripts)
		descriptorMap := indexFilesByTeamProject(descriptorFiles)
		sidecarMap := indexSidecarsByTeamProject(sidecarFiles)

		for k, _ := range buildScriptMap {
			if _, present := descriptorMap[k]; !present {
				Log.Printf("Skipping project without a descriptor: %s\n", k)
				continue
			}

			descriptorData, err := ioutil.ReadFile(descriptorMap[k])
			if err != nil {
				Log.Println(err)
				continue
			}

			descriptor, err := descriptorForTeamProject(descriptorData)
			if err != nil {
				Log.Println(err)
				continue
			}
			if descriptor.Image == "" {
				Log.Printf("Skipping project %s without descriptor build image: %+v\n", k, descriptor)
				continue
			}

			sidecars := readSidecars(sidecarMap[k])

			parts := strings.Split(k, "/")
			p := Project{
				Team:        parts[0],
				ProjectName: parts[1],
				Descriptor:  descriptor,
				Sidecars:    sidecars,
			}
			projects[k] = p
		}
		return nil
	}

	err := retry.New(32*time.Second, 5, func(attempts uint) {
		if attempts == 0 {
			return
		}
		Log.Printf("Wait for clone-repository with-backoff try %d\n", attempts+1)
		time.Sleep((1 << attempts) * time.Second)
	}).Try(work)

	if err != nil {
		return nil, err
	}
	return projects, nil
}

func getProjects() map[string]Project {
	p := <-projectGetChan
	pm := make(map[string]Project, len(p))
	for k, v := range p {
		pm[k] = v
	}
	return pm
}

func setProjects(p map[string]Project) {
	projectSetChan <- p
}

func projectByTeamName(team, project string) (Project, bool) {
	projects := getProjects()
	key := projectKey(team, project)
	p, ok := projects[key]
	return p, ok
}

func projectKey(team, project string) string {
	return fmt.Sprintf("%s/%s", team, project)
}

func teamProject(file string) (string, string, error) {
	parts := strings.Split(file, "/")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("Path does not contain minimum depth of 3: %s", file)
	}
	return parts[len(parts)-3], parts[len(parts)-2], nil
}

func indexFilesByTeamProject(files []string) map[string]string {
	m := make(map[string]string)
	for _, file := range files {
		team, project, err := teamProject(file)
		if err != nil {
			Log.Println(err)
			continue
		}
		key := projectKey(team, project)
		m[key] = file
	}
	return m
}

func indexSidecarsByTeamProject(files []string) map[string][]string {
	m := make(map[string][]string)
	for _, file := range files {
		team, project, err := teamProject(file)
		if err != nil {
			Log.Println(err)
			continue
		}
		key := projectKey(team, project)
		arr, present := m[key]
		if !present {
			arr = make([]string, 0)
		}
		arr = append(arr, file)
		m[key] = arr
	}
	return m
}

func readSidecars(files []string) []string {
	arr := make([]string, len(files))
	for i, v := range files {
		data, err := ioutil.ReadFile(v)
		if err != nil {
			Log.Println(err)
			continue
		}
		arr[i] = string(data)
	}
	return arr
}

func descriptorForTeamProject(data []byte) (ProjectDescriptor, error) {
	var descriptor ProjectDescriptor
	if err := json.Unmarshal(data, &descriptor); err != nil {
		return ProjectDescriptor{}, err
	}

	if descriptor.ManagedRefRegexStr != "" {
		if re, err := regexp.Compile(descriptor.ManagedRefRegexStr); err != nil {
			Log.Printf("Error parsing managed-branch-regex %s for descriptor %+v: %v\n", descriptor.ManagedRefRegexStr, data, err)
		} else {
			descriptor.regex = re
		}
	}

	return descriptor, nil
}
