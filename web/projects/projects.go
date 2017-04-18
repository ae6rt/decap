package projects

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/retry"
	"github.com/pkg/errors"
)

const buildScriptRegex = `build\.sh`
const projectDescriptorRegex = `project\.json`
const sideCarRegex = `^.+-sidecar\.json`

var projectsView map[string]v1.Project

// DefaultProjectManager is the live, working projects manager that clones the build scripts repo and returns associated information.
type DefaultProjectManager struct {
	repo   string
	branch string
	rwLock *sync.RWMutex
	logger *log.Logger
}

// NewDefaultManager returns the working project manager.
func NewDefaultManager(repo, branch string, logger *log.Logger) ProjectManager {
	return DefaultProjectManager{repo: repo, branch: branch, logger: logger, rwLock: &sync.RWMutex{}}
}

// Assemble assembles the build scripts repo into a manageable API.
func (t DefaultProjectManager) Assemble() error {
	t.rwLock.Lock()
	defer t.rwLock.Unlock()

	var err error
	projectsView, err = t.assembleProjects()
	if err != nil {
		return errors.Wrap(err, "Error assembling project")
	}
	return nil
}

// GetProjects returns current in-memory view of all projects.
func (t DefaultProjectManager) GetProjects() map[string]v1.Project {
	t.rwLock.RLock()
	defer t.rwLock.RUnlock()

	pm := make(map[string]v1.Project, len(projectsView))
	for k, v := range projectsView {
		pm[k] = v
	}
	return pm
}

// Get returns the project by key.
func (t DefaultProjectManager) Get(projectKey string) *v1.Project {
	t.rwLock.RLock()
	defer t.rwLock.RUnlock()

	p, ok := projectsView[projectKey]
	if !ok {
		return nil
	}

	// todo make a copy first
	return &p
}

// GerProjectByTeamName retrieves a project indexed by Team and Project name.
func (t DefaultProjectManager) GetProjectByTeamName(team, project string) (v1.Project, bool) {
	t.rwLock.RLock()
	defer t.rwLock.RUnlock()

	key := projectKey(team, project)
	p, ok := projectsView[key]
	return p, ok
}

// RepositoryURL returns the URL of the underlying projects repository.
func (t DefaultProjectManager) RepositoryURL() string {
	return t.repo
}

// RepositoryBranch returns the branch of the underlying repository.
func (t DefaultProjectManager) RepositoryBranch() string {
	return t.branch
}

func (t DefaultProjectManager) assembleProjects() (map[string]v1.Project, error) {
	projects := make(map[string]v1.Project, 0)
	work := func() error {
		t.logger.Printf("Clone build-scripts repository...\n")
		cloneDirectory, err := ioutil.TempDir("", "repoclone-")
		defer func() {
			if err := os.RemoveAll(cloneDirectory); err != nil {
				t.logger.Printf("assembleProjects(%s,%s) error removing clone directory %s: %v\n", t.repo, t.branch, cloneDirectory, err)
			}
		}()

		if err != nil {
			return err
		}
		if err := t.clone(t.repo, t.branch, cloneDirectory, true); err != nil {
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

		buildScriptMap := t.indexFilesByTeamProject(buildScripts)
		descriptorMap := t.indexFilesByTeamProject(descriptorFiles)
		sidecarMap := t.indexSidecarsByTeamProject(sidecarFiles)

		for k, _ := range buildScriptMap {
			if _, present := descriptorMap[k]; !present {
				t.logger.Printf("Skipping project without a descriptor: %s\n", k)
				continue
			}

			descriptorData, err := ioutil.ReadFile(descriptorMap[k])
			if err != nil {
				t.logger.Println(err)
				continue
			}

			descriptor, err := t.descriptorForTeamProject(descriptorData)
			if err != nil {
				t.logger.Println(err)
				continue
			}
			if descriptor.Image == "" {
				t.logger.Printf("Skipping project %s without descriptor build image: %+v\n", k, descriptor)
				continue
			}

			sidecars := t.readSidecars(sidecarMap[k])

			parts := strings.Split(k, "/")
			p := v1.Project{
				Team:        parts[0],
				ProjectName: parts[1],
				Descriptor:  descriptor,
				Sidecars:    sidecars,
			}
			projects[k] = p
		}
		return nil
	}

	err := retry.New(5, func(attempts int) {
		if attempts == 0 {
			return
		}
		t.logger.Printf("Wait for clone-repository with-backoff try %d\n", attempts+1)
		time.Sleep((1 << uint(attempts)) * time.Second)
	}).Try(work)

	if err != nil {
		return nil, err
	}
	return projects, nil
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

func (t DefaultProjectManager) indexFilesByTeamProject(files []string) map[string]string {
	m := make(map[string]string)
	for _, file := range files {
		team, project, err := teamProject(file)
		if err != nil {
			t.logger.Println(err)
			continue
		}
		key := projectKey(team, project)
		m[key] = file
	}
	return m
}

func (t DefaultProjectManager) indexSidecarsByTeamProject(files []string) map[string][]string {
	m := make(map[string][]string)
	for _, file := range files {
		team, project, err := teamProject(file)
		if err != nil {
			t.logger.Println(err)
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

func (t DefaultProjectManager) readSidecars(files []string) []string {
	arr := make([]string, len(files))
	for i, v := range files {
		data, err := ioutil.ReadFile(v)
		if err != nil {
			t.logger.Println(err)
			continue
		}
		arr[i] = string(data)
	}
	return arr
}

func (t DefaultProjectManager) descriptorForTeamProject(data []byte) (v1.ProjectDescriptor, error) {
	var descriptor v1.ProjectDescriptor
	if err := json.Unmarshal(data, &descriptor); err != nil {
		return v1.ProjectDescriptor{}, err
	}

	if descriptor.ManagedRefRegexStr != "" {
		if re, err := regexp.Compile(descriptor.ManagedRefRegexStr); err != nil {
			t.logger.Printf("Error parsing managed-branch-regex %s for descriptor %+v: %v\n", descriptor.ManagedRefRegexStr, data, err)
		} else {
			descriptor.Regex = re
		}
	}

	return descriptor, nil
}

// Clone clones the Git repository at repositoryURL at the given branch into the given directory.  If shallow is true, use --depth 1.
func (t DefaultProjectManager) clone(repositoryURL, branch, dir string, shallow bool) error {
	args := []string{"clone"}
	if shallow {
		args = append(args, []string{"--depth", "1"}...)
	}
	args = append(args, []string{"--branch", branch, repositoryURL, dir}...)
	return t.executeShellCommand("git", args)
}

func (t DefaultProjectManager) executeShellCommand(commandName string, args []string) error {
	t.logger.Printf("Executing %s %+v\n", commandName, args)
	command := exec.Command(commandName, args...)
	var stdOutErr []byte
	var err error
	stdOutErr, err = command.CombinedOutput()
	if err != nil {
		return err
	}
	t.logger.Printf("%v\n", string(stdOutErr))

	return nil
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

	var files []string
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
