package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ae6rt/gittools"
	"github.com/ae6rt/retry"
)

const buildScriptRegex = `build\.sh`
const projectDescriptorRegex = `project\.json`
const sideCarRegex = `^.+-sidecar\.json`

var projects map[string]Project
var projectMutex = &sync.Mutex{}

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
			files = append(files, root+"/"+path)
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

func findBuildScripts(root string) ([]string, error) {
	files := make([]string, 0)
	markFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(path, ".git") {
			return filepath.SkipDir
		}

		// record this as a too-deep path we never want to traverse again
		if info.IsDir() && strings.Count(path, "/") > 2 {
			return filepath.SkipDir
		}

		// We compare again the depth to 2 because ./build.sh is an undesirable possibility.
		if strings.Count(path, "/") == 2 && info.Mode().IsRegular() && info.Name() == "build.sh" {
			files = append(files, root+"/"+path)
		}
		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.Chdir(pwd)
		if err != nil {
			Log.Printf("findBuildScripts cannot restore working directory to %s: %v\n", pwd, err)
		}
	}()

	// Change directory to root so we have no need to know how many "/" root itself contains.
	if err := os.Chdir(root); err != nil {
		return nil, err
	}

	// Walk relative to root.
	err = filepath.Walk(".", markFn)
	if err != nil {
		return nil, err
	}

	return files, nil
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

func descriptorForTeamProject(file string) (ProjectDescriptor, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return ProjectDescriptor{}, err
	}
	var descriptor ProjectDescriptor
	if err := json.Unmarshal(data, &descriptor); err != nil {
		return ProjectDescriptor{}, err
	}
	return descriptor, nil
}

func assembleProjects(scriptsRepo, scriptsRepoBranch string) (map[string]Project, error) {
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

		// The build scripts are considered the anchor for a project.
		// If the build.sh does not exist, then the project effectively does not exist.
		// We could have chosen the project.json descriptor, as this must also exist for a valid project.
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
			_, present := descriptorMap[k]
			if !present {
				Log.Printf("Skipping project without a descriptor: %s\n", k)
				continue
			}
			descriptor, err := descriptorForTeamProject(descriptorMap[k])
			if err != nil {
				Log.Println(err)
				continue
			}

			sidecars := readSidecars(sidecarMap[k])

			parts := strings.Split(k, "/")
			p := Project{
				Parent:     parts[0],
				Library:    parts[1],
				Descriptor: descriptor,
				Sidecars:   sidecars,
			}
			projects[k] = p
		}
		return nil
	}

	err := retry.New(5*time.Second, 10, retry.DefaultBackoffFunc).Try(work)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

/*
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
*/

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
