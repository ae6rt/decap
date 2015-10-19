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

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

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
	version := v1.Version{
		Version: buildVersion,
		Commit:  buildCommit,
		Date:    buildDate,
		SDK:     buildGoSDK,
	}
	w.Header().Set("Content-type", "application/json")

	data, err := json.Marshal(&version)
	if err != nil {
		version = v1.Version{Meta: v1.Meta{Error: err.Error()}}
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

	a := make([]v1.Team, 0)
	for k, _ := range keys {
		a = append(a, v1.Team{Name: k})
	}

	w.Header().Set("Content-type", "application/json")

	teams := v1.Teams{Teams: a}
	data, err := json.Marshal(&teams)
	if err != nil {
		teams := v1.Teams{Meta: v1.Meta{Error: err.Error()}}
		data, _ := json.Marshal(&teams)
		w.WriteHeader(500)
		w.Write(data)
		return
	}
	w.Write(data)
}

func ProjectsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	team := r.URL.Query().Get("team")
	arr := make([]v1.Project, 0)
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

	p := v1.Projects{Projects: arr}
	data, err := json.Marshal(&p)
	if err != nil {
		p := v1.Projects{Meta: v1.Meta{Error: err.Error()}}
		data, _ := json.Marshal(&p)
		w.WriteHeader(500)
		w.Write(data)
		return
	}
	w.Write(data)
}

func DeferredBuildsHandler(builder Builder) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		switch r.Method {
		case "GET":
			deferred, err := builder.DeferredBuilds()
			if err != nil {
				w.WriteHeader(500)
				w.Write(simpleError(err))
				return
			}
			squashed, _ := builder.SquashDeferred(deferred)
			var v v1.Deferred
			v.DeferredEvents = squashed
			data, err := json.Marshal(&v)
			if err != nil {
				w.WriteHeader(500)
				w.Write(simpleError(err))
				return
			}
			w.Write(data)
		case "POST":
			// todo implement delete deferred build
			w.WriteHeader(400)
			w.Write(simpleError(fmt.Errorf("Unsupported method: %s", r.Method)))
			return
		default:
			w.WriteHeader(400)
			w.Write(simpleError(fmt.Errorf("Unsupported method: %s", r.Method)))
			return
		}
	}
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

		event := v1.UserBuildEvent{Team_: team, Project_: project, Refs_: branches}
		go decap.LaunchBuild(event)
	}
}

func simpleError(err error) []byte {
	m := v1.Meta{Error: err.Error()}
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
				data, _ := json.Marshal(&v1.Refs{Meta: v1.Meta{Error: err.Error()}})
				w.WriteHeader(500)
				w.Write(data)
				return
			}

			branches := v1.Refs{Refs: nativeBranches}
			data, err := json.Marshal(&branches)
			if err != nil {
				Log.Print(err)
				data, _ := json.Marshal(&v1.Refs{Meta: v1.Meta{Error: err.Error()}})
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
			builds := v1.Builds{Meta: v1.Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(400)
			w.Write(data)
			return
		}

		limit, err := toUint64(r.URL.Query().Get("limit"), 100)
		if err != nil {
			builds := v1.Builds{Meta: v1.Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(400)
			w.Write(data)
			return
		}

		buildList, err := storageService.GetBuildsByProject(v1.Project{Team: team, ProjectName: project}, since, limit)

		if err != nil {
			builds := v1.Builds{Meta: v1.Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(502)
			w.Write(data)
			return
		}

		builds := v1.Builds{Builds: buildList}
		data, err := json.Marshal(&builds)
		if err != nil {
			builds := v1.Builds{Meta: v1.Meta{Error: err.Error()}}
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
	switch r.Method {
	case "POST":
		state := params.ByName("state")

		shutdownState := v1.Shutdown(state)
		switch shutdownState {
		case BUILD_QUEUE_CLOSE:
			if <-getShutdownChan == BUILD_QUEUE_OPEN {
				Log.Printf("Shutdown state changed to %s\n", state)
			}
			setShutdownChan <- shutdownState
		case BUILD_QUEUE_OPEN:
			if <-getShutdownChan == BUILD_QUEUE_CLOSE {
				Log.Printf("Shutdown state changed to %s\n", state)
			}
			setShutdownChan <- shutdownState
		default:
			w.WriteHeader(400)
			w.Write(simpleError(fmt.Errorf("Unsupported shutdown state: %v", shutdownState)))
			return
		}
	case "GET":
		var data []byte
		data, _ = json.Marshal(&v1.ShutdownState{State: <-getShutdownChan})
		w.Write(data)
	}
}

func LogLevelHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	switch r.Method {
	case "POST":
		level := params.ByName("level")

		logLevel := LogLevel(level)
		switch logLevel {
		case LOG_DEFAULT:
			if <-getLogLevelChan == LOG_DEBUG {
				Log.Printf("Log level changed to %s\n", level)
			}
			setLogLevelChan <- logLevel
		case LOG_DEBUG:
			if <-getLogLevelChan == LOG_DEFAULT {
				Log.Printf("Shutdown state changed to %s\n", level)
			}
			setLogLevelChan <- logLevel
		default:
			w.WriteHeader(400)
			w.Write(simpleError(fmt.Errorf("Unsupported log level: %v", logLevel)))
			return
		}
		w.WriteHeader(200)
		return
	}
}
