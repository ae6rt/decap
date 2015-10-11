package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/julienschmidt/httprouter"
)

var shutdownMutex = &sync.Mutex{}
var shutdown bool

func toUint64(value string, dflt uint64) (uint64, error) {
	if value == "" {
		return dflt, nil
	}
	if i, err := strconv.ParseUint(value, 10, 64); err != nil {
		return 0, err
	} else {
		return i, nil
	}
}

func VersionHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	version := Version{
		Version: buildVersion,
		Commit:  buildCommit,
		Date:    buildDate,
		SDK:     buildGoSDK,
	}
	w.Header().Set("Content-type", "application/json")

	data, err := json.Marshal(&version)
	if err != nil {
		version = Version{Meta: Meta{Error: err.Error()}}
		data, _ := json.Marshal(&version)
		w.WriteHeader(500)
		w.Write(data)
		return
	}
	w.Write(data)
}

func TeamsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	p := getProjects()

	keys := make(map[string]string)
	for _, v := range p {
		keys[v.Team] = ""
	}

	a := make([]Team, 0)
	for k, _ := range keys {
		a = append(a, Team{Name: k})
	}

	w.Header().Set("Content-type", "application/json")

	teams := Teams{Teams: a}
	data, err := json.Marshal(&teams)
	if err != nil {
		teams := Teams{Meta: Meta{Error: err.Error()}}
		data, _ := json.Marshal(&teams)
		w.WriteHeader(500)
		w.Write(data)
		return
	}
	w.Write(data)
}

func ProjectsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	team := r.URL.Query().Get("team")
	arr := make([]Project, 0)
	if team != "" {
		for _, v := range getProjects() {
			if team == v.Team {
				arr = append(arr, v)
			}
		}
	} else {
		for _, v := range getProjects() {
			arr = append(arr, v)
		}
	}

	w.Header().Set("Content-type", "application/json")

	p := Projects{Projects: arr}
	data, err := json.Marshal(&p)
	if err != nil {
		p := Projects{Meta: Meta{Error: err.Error()}}
		data, _ := json.Marshal(&p)
		w.WriteHeader(500)
		w.Write(data)
		return
	}
	w.Write(data)
}

func ExecuteBuildHandler(decap Builder) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		team := params.ByName("team")
		project := params.ByName("project")

		if _, present := projectByTeamName(team, project); !present {
			Log.Printf("Unknown project %s/%s", team, project)
			w.WriteHeader(404)
			w.Write(simpleError(fmt.Errorf("Unknown project %s/%s", team, project)))
			return
		}

		branches := r.URL.Query()["branch"]
		if len(branches) == 0 {
			Log.Println("No branches specified")
			w.WriteHeader(400)
			w.Write(simpleError(fmt.Errorf("No branches specified")))
			return
		}

		event := UserBuildEvent{Team_: team, Project_: project, Refs_: branches}
		go decap.LaunchBuild(event)
	}
}

func simpleError(err error) []byte {
	m := Meta{Error: err.Error()}
	data, _ := json.Marshal(&m)
	return data
}

func HooksHandler(buildScriptsRepo, buildScriptsBranch string, decap Builder) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		repoManager := params.ByName("repomanager")

		if r.Body == nil {
			Log.Println("Expecting an HTTP entity")
			w.WriteHeader(400)
			w.Write(simpleError(fmt.Errorf("Expecting an HTTP entity")))
			return
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Log.Println(err)
			w.WriteHeader(500)
			w.Write(simpleError(err))
			return
		}
		defer func() {
			r.Body.Close()
		}()

		Log.Printf("%s hook received: %s\n", repoManager, data)

		switch repoManager {
		case "buildscripts":
			if p, err := assembleProjects(buildScriptsRepo, buildScriptsBranch); err != nil {
				w.WriteHeader(500)
				w.Write(simpleError(err))
			} else {
				setProjects(p)
				Log.Println("Build scripts refreshed via post commit hook")
			}
		case "github":
			eventType := r.Header.Get("X-Github-Event")
			switch eventType {
			case "create":
				event := GithubEvent{}
				if err := json.Unmarshal(data, &event); err != nil {
					Log.Println(err)
					w.WriteHeader(500)
					w.Write(simpleError(err))
					return
				}
				go decap.LaunchBuild(event)
			case "push":
				event := GithubEvent{}
				if err := json.Unmarshal(data, &event); err != nil {
					Log.Println(err)
					w.WriteHeader(500)
					w.Write(simpleError(err))
					return
				}
				go decap.LaunchBuild(event)
			default:
				w.WriteHeader(400)
				w.Write(simpleError(fmt.Errorf("Github hook missing event type header.  See https://developer.github.com/webhooks/#delivery-headers.")))
				return
			}
		default:
			Log.Printf("repomanager %s not supported\n", repoManager)
			w.WriteHeader(400)
			w.Write(simpleError(fmt.Errorf("repomanager %s not supported", repoManager)))
			return
		}
		w.WriteHeader(200)
	}
}

// todo Cleanup with PreStop might help explain to the world the state of the container immediately before termination.
// See https://godoc.org/github.com/kubernetes/kubernetes/pkg/api/v1#Lifecycle
func StopBuildHandler(decap Builder) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		buildID := params.ByName("id")
		if err := decap.DeletePod(buildID); err != nil {
			Log.Println(err)
			w.WriteHeader(500)
			w.Write(simpleError(err))
		}
	}
}

// Handle requests for branches and tags on a project
func ProjectRefsHandler(repoClients map[string]SCMClient) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		team := params.ByName("team")
		projectName := params.ByName("project")

		project, present := projectByTeamName(team, projectName)
		if !present {
			w.WriteHeader(404)
			w.Write(simpleError(fmt.Errorf("Unknown project %s/%s", team, projectName)))
			return
		}

		repositoryManager := project.Descriptor.RepoManager
		switch repositoryManager {
		case "github":
			w.Header().Set("Content-type", "application/json")
			repoClient := repoClients["github"]
			nativeBranches, err := repoClient.GetRefs(project.Team, project.ProjectName)
			if err != nil {
				Log.Print(err)
				data, _ := json.Marshal(&Refs{Meta: Meta{Error: err.Error()}})
				w.WriteHeader(500)
				w.Write(data)
				return
			}

			branches := Refs{Refs: nativeBranches}
			data, err := json.Marshal(&branches)
			if err != nil {
				Log.Print(err)
				data, _ := json.Marshal(&Refs{Meta: Meta{Error: err.Error()}})
				w.WriteHeader(500)
				w.Write(data)
				return
			}

			w.Write(data)
			return
		default:
			w.WriteHeader(400)
			w.Write(simpleError(fmt.Errorf("repomanager not supported: %s", repositoryManager)))
		}
	}
}

// Return gzipped console log, or console log in plain text if Accept: text/plain is set
func LogHandler(storageService StorageService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		buildID := params.ByName("id")
		data, err := storageService.GetConsoleLog(buildID)
		if err != nil {
			Log.Println(err)
			w.WriteHeader(500)
			w.Write(simpleError(err))
			return
		}
		if r.Header.Get("Accept") == "text/plain" {
			inputBuffer := bytes.NewBuffer(data)
			source, err := gzip.NewReader(inputBuffer)
			if err != nil {
				Log.Println(err)
				w.WriteHeader(500)
				w.Write(simpleError(err))
				return
			}
			defer func() {
				source.Close()
			}()

			var outputBuffer = new(bytes.Buffer)
			io.Copy(outputBuffer, source)
			w.Header().Set("Content-type", "text/plain")
			w.Write(outputBuffer.Bytes())

		} else {
			w.Header().Set("Content-type", "application/x-gzip")
			w.Write(data)
		}
	}
}

// Return artifacts gzipped tarball, or file listing in tarball if Accept: text/plain is set
func ArtifactsHandler(storageService StorageService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		buildID := params.ByName("id")
		data, err := storageService.GetArtifacts(buildID)
		if err != nil {
			Log.Println(err)
			w.WriteHeader(500)
			w.Write(simpleError(err))
			return
		}
		if r.Header.Get("Accept") == "text/plain" {
			// unzip
			inputBuffer := bytes.NewBuffer(data)
			source, err := gzip.NewReader(inputBuffer)
			if err != nil {
				Log.Println(err)
				w.WriteHeader(500)
				w.Write(simpleError(err))
				return
			}
			defer func() {
				source.Close()
			}()

			var outputBuffer = new(bytes.Buffer)
			io.Copy(outputBuffer, source)

			newData := outputBuffer.Bytes()

			// get tar manifest
			plainReader := bytes.NewReader(newData)
			tr := tar.NewReader(plainReader)
			records := ""
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					Log.Println(err)
					w.WriteHeader(500)
					w.Write(simpleError(err))
					return
				}
				records = fmt.Sprintf("%s\n%s", records, hdr.Name)
			}
			w.Header().Set("Content-type", "text/plain")
			w.Write([]byte(records))
		} else {
			w.Header().Set("Content-type", "application/x-gzip")
			w.Write(data)
		}
	}
}

func BuildsHandler(storageService StorageService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		team := params.ByName("team")
		project := params.ByName("project")

		since, err := toUint64(r.URL.Query().Get("since"), 0)
		if err != nil {
			builds := Builds{Meta: Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(400)
			w.Write(data)
			return
		}

		limit, err := toUint64(r.URL.Query().Get("limit"), 100)
		if err != nil {
			builds := Builds{Meta: Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(400)
			w.Write(data)
			return
		}

		buildList, err := storageService.GetBuildsByProject(Project{Team: team, ProjectName: project}, since, limit)

		if err != nil {
			builds := Builds{Meta: Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(502)
			w.Write(data)
			return
		}

		builds := Builds{Builds: buildList}
		data, err := json.Marshal(&builds)
		if err != nil {
			builds := Builds{Meta: Meta{Error: err.Error()}}
			builds.Meta.Error = err.Error()
			data, _ := json.Marshal(&builds)
			w.WriteHeader(500)
			w.Write(data)
			return
		}
		w.Write(data)
	}
}

func ShutdownHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	state := params.ByName("state")
	switch state {
	case "close":
		shutdownMutex.Lock()
		defer shutdownMutex.Unlock()
		shutdown = true
	case "open":
		shutdownMutex.Lock()
		defer shutdownMutex.Unlock()
		shutdown = false
	}
	w.WriteHeader(200)
}
