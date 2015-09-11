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

		buildScripts, err := filesByRegex(cloneDirectory, buildScriptRegex)
		if err != nil {
			return err
		}
		descriptors, err := filesByRegex(cloneDirectory, projectDescriptorRegex)
		if err != nil {
			return err
		}
		sidecars, err := filesByRegex(cloneDirectory, sideCarRegex)
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

// deprecated in favor of assembleProjects()
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
