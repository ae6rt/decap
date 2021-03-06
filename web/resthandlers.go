package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/projects"
	"github.com/ae6rt/decap/web/scmclients"
	"github.com/ae6rt/decap/web/storage"
	"github.com/julienschmidt/httprouter"
)

func toUint64(value string, dflt uint64) (uint64, error) {
	if value == "" {
		return dflt, nil
	}
	return strconv.ParseUint(value, 10, 64)
}

// VersionHandler returns the decap server information.
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
		_, _ = w.Write(data)
		return
	}

	_, _ = w.Write(data)
}

// Teamster
func TeamsHandler(projectManager projects.ProjectManager) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		p := projectManager.GetProjects()

		keys := make(map[string]struct{})
		for _, v := range p {
			keys[v.Team] = struct{}{}
		}

		var a []v1.Team
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
			_, _ = w.Write(data)
			return
		}

		_, _ = w.Write(data)
	}
}

// ProjectsHandler returns informtion about managed projects.
func ProjectsHandler(projectManager projects.ProjectManager) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-type", "application/json")

		team := r.URL.Query().Get("team")

		allProjects := projectManager.GetProjects()

		var arr []v1.Project
		if team != "" {
			for _, v := range allProjects {
				if team == v.Team {
					arr = append(arr, v)
				}
			}
		} else {
			for _, v := range allProjects {
				arr = append(arr, v)
			}
		}

		p := v1.Projects{Projects: arr}
		data, err := json.Marshal(&p)
		if err != nil {
			p := v1.Projects{Meta: v1.Meta{Error: err.Error()}}
			data, _ := json.Marshal(&p)
			w.WriteHeader(500)
			_, _ = w.Write(data)
			return
		}

		_, _ = w.Write(data)
	}
}

// DeferredBuildsHandler returns information about deferred builds.
func DeferredBuildsHandler(buildManager BuildManager) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		w.Header().Set("Content-type", "application/json")

		switch r.Method {
		case "GET":
			deferred, err := buildManager.DeferredBuilds()
			if err != nil {
				w.WriteHeader(500)
				_, _ = w.Write(simpleError(err))
				return
			}

			data, err := json.Marshal(&deferred)
			if err != nil {
				w.WriteHeader(500)
				_, _ = w.Write(simpleError(err))
				return
			}
			_, _ = w.Write(data)
		case "POST":
			key := r.URL.Query().Get("key")
			if key == "" {
				w.WriteHeader(400)
				_, _ = w.Write(simpleError(fmt.Errorf("Missing or empty key parameter in clear deferred build.")))
				return
			}
			if err := buildManager.ClearDeferredBuild(key); err != nil {
				w.WriteHeader(500)
				_, _ = w.Write(simpleError(err))
			}
		}
	}
}

// ExecuteBuildHandler handles user-requested build executions.
func ExecuteBuildHandler(buildManager BuildManager, logger *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		team := params.ByName("team")
		project := params.ByName("project")

		branches := r.URL.Query()["branch"]

		if len(branches) == 0 {
			logger.Println("No branches specified.")
			w.WriteHeader(400)
			_, _ = w.Write(simpleError(fmt.Errorf("No branches specified")))
			return
		}

		for _, b := range branches {
			event := v1.UserBuildEvent{Team: team, Project: project, Ref: b}
			go func() {
				if err := buildManager.LaunchBuild(event); err != nil {
					logger.Printf("Error launching build: %v\n", err)
				}
			}()
		}
		w.WriteHeader(200)
	}
}

// HooksHandler handles externally originated SCM events that trigger builds or build-scripts repository refreshes.
func HooksHandler(projectManager projects.ProjectManager, buildManager BuildManager, logger *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		repoManager := params.ByName("repomanager")

		if r.Body == nil {
			logger.Println("Expecting an HTTP entity")
			w.WriteHeader(400)
			return
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Printf("Error reading hook payload: %v\n", err)
			w.WriteHeader(400)
			return
		}
		defer func() {
			_ = r.Body.Close()
		}()

		switch repoManager {
		case "buildscripts":
			err := projectManager.Assemble()
			if err != nil {
				logger.Printf("Error refreshing build scripts: %v\n", err)
				w.WriteHeader(500)
				return
			}
			logger.Println("Build scripts refreshed.")
		case "github":
			var event GithubEvent

			eventType := r.Header.Get("X-Github-Event")
			switch eventType {
			case "create":
			case "push":
			default:
				logger.Printf("Unhandled Github event type <%s>.  See https://developer.github.com/webhooks/#delivery-headers.", eventType)
				w.WriteHeader(400)
				return
			}

			if err := json.Unmarshal(data, &event); err != nil {
				logger.Printf("Error unmarshaling Github event: %v\n", err)
				w.WriteHeader(500)
				return
			}

			go func() {
				if err := buildManager.LaunchBuild(event.BuildEvent()); err != nil {
					logger.Printf("Error launching build for event %+v: %v\n", event, err)
					w.WriteHeader(500)
					return
				}
			}()
		default:
			logger.Printf("repomanager %s not supported\n", repoManager)
			w.WriteHeader(400)
			return
		}

		w.WriteHeader(200)
	}
}

// StopBuildHandler deletes the pod executing the specified build ID.
// todo inject a logger
func StopBuildHandler(buildManager BuildManager, logger *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		w.Header().Set("Content-type", "application/json")

		buildID := params.ByName("id")
		if err := buildManager.DeletePod(buildID); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write(simpleError(err))
			return
		}
	}
}

// ProjectRefsHandler handles informational requests for branches and tags on a project
func ProjectRefsHandler(projectManager projects.ProjectManager, repoClients map[string]scmclients.SCMClient, logger *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		w.Header().Set("Content-type", "application/json")

		team := params.ByName("team")
		projectName := params.ByName("project")

		project, present := projectManager.GetProjectByTeamName(team, projectName)

		if !present {
			w.WriteHeader(404)
			_, _ = w.Write(simpleError(fmt.Errorf("Unknown project %s/%s", team, projectName)))
			return
		}

		repositoryManager := project.Descriptor.RepoManager

		switch repositoryManager {
		case "github":
			repoClient := repoClients["github"]
			nativeBranches, err := repoClient.GetRefs(project.Team, project.ProjectName)
			if err != nil {
				logger.Printf("Error retrieving refs for %s/%s: %v\n", project.Team, project.ProjectName, err)
				data, _ := json.Marshal(&v1.Refs{Meta: v1.Meta{Error: err.Error()}})
				w.WriteHeader(500)
				_, _ = w.Write(data)
				return
			}

			branches := v1.Refs{Refs: nativeBranches}
			data, err := json.Marshal(&branches)
			if err != nil {
				logger.Printf("Error serializing native branches for %s/%s: %v\n", project.Team, project.ProjectName, err)
				data, _ := json.Marshal(&v1.Refs{Meta: v1.Meta{Error: err.Error()}})
				w.WriteHeader(500)
				_, _ = w.Write(data)
				return
			}
			_, _ = w.Write(data)
		default:
			w.WriteHeader(400)
			_, _ = w.Write(simpleError(fmt.Errorf("repository manager for %s/%s not supported: %s", project.Team, project.ProjectName, repositoryManager)))
		}
	}
}

// Return gzipped console log, or console log in plain text if Accept: text/plain is set
func LogHandler(buildStore storage.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		buildID := params.ByName("id")
		data, err := buildStore.GetConsoleLog(buildID)
		if err != nil {
			Log.Println(err)
			w.WriteHeader(500)
			_, _ = w.Write(simpleError(err))
			return
		}
		if r.Header.Get("Accept") == "text/plain" {
			inputBuffer := bytes.NewBuffer(data)
			source, err := gzip.NewReader(inputBuffer)
			if err != nil {
				Log.Println(err)
				w.WriteHeader(500)
				_, _ = w.Write(simpleError(err))
				return
			}
			defer func() {
				_ = source.Close()
			}()

			var outputBuffer = new(bytes.Buffer)
			_, _ = io.Copy(outputBuffer, source)
			w.Header().Set("Content-type", "text/plain")
			_, _ = w.Write(outputBuffer.Bytes())

		} else {
			w.Header().Set("Content-type", "application/x-gzip")
			_, _ = w.Write(data)
		}
	}
}

// ArtifactsHandler returns build artifacts gzipped tarball, or file listing in tarball if Accept: text/plain is set
func ArtifactsHandler(buildStore storage.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		buildID := params.ByName("id")
		data, err := buildStore.GetArtifacts(buildID)
		if err != nil {
			Log.Println(err)
			w.WriteHeader(500)
			_, _ = w.Write(simpleError(err))
			return
		}
		if r.Header.Get("Accept") == "text/plain" {
			// unzip
			inputBuffer := bytes.NewBuffer(data)
			source, err := gzip.NewReader(inputBuffer)
			if err != nil {
				Log.Println(err)
				w.WriteHeader(500)
				_, _ = w.Write(simpleError(err))
				return
			}
			defer func() {
				_ = source.Close()
			}()

			var outputBuffer = new(bytes.Buffer)
			_, _ = io.Copy(outputBuffer, source)

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
					_, _ = w.Write(simpleError(err))
					return
				}
				records = fmt.Sprintf("%s\n%s", records, hdr.Name)
			}
			w.Header().Set("Content-type", "text/plain")
			_, _ = w.Write([]byte(records))
		} else {
			w.Header().Set("Content-type", "application/x-gzip")
			_, _ = w.Write(data)
		}
	}
}

// BuildsHandler handles requests for historical build info.
func BuildsHandler(buildStore storage.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		team := params.ByName("team")
		project := params.ByName("project")

		w.Header().Set("Content-type", "application/json")

		since, err := toUint64(r.URL.Query().Get("since"), 0)
		if err != nil {
			builds := v1.Builds{Meta: v1.Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(400)
			_, _ = w.Write(data)
			return
		}

		limit, err := toUint64(r.URL.Query().Get("limit"), 100)
		if err != nil {
			builds := v1.Builds{Meta: v1.Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(400)
			_, _ = w.Write(data)
			return
		}

		buildList, err := buildStore.GetBuildsByProject(v1.Project{Team: team, ProjectName: project}, since, limit)

		if err != nil {
			builds := v1.Builds{Meta: v1.Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(502)
			_, _ = w.Write(data)
			return
		}

		builds := v1.Builds{Builds: buildList}
		data, err := json.Marshal(&builds)
		if err != nil {
			builds := v1.Builds{Meta: v1.Meta{Error: err.Error()}}
			builds.Meta.Error = err.Error()
			data, _ := json.Marshal(&builds)
			w.WriteHeader(500)
			_, _ = w.Write(data)
			return
		}
		_, _ = w.Write(data)
	}
}

// ShutdownHandler
func ShutdownHandler(buildManager BuildManager) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		switch r.Method {
		case "POST":
			shutdownState := params.ByName("state")
			switch shutdownState {
			case BuildQueueClose:
				buildManager.OpenQueue()
			case BuildQueueOpen:
				buildManager.CloseQueue()
			default:
				w.Header().Set("Content-type", "application/json")
				w.WriteHeader(400)
				_, _ = w.Write(simpleError(fmt.Errorf("Unsupported shutdown state: %v.  Valid states are [%s|%s]", shutdownState, BuildQueueOpen, BuildQueueClose)))
				return
			}
		case "GET":
			var state string

			switch buildManager.QueueIsOpen() {
			case true:
				state = BuildQueueOpen
			case false:
				state = BuildQueueClose
			}
			var data []byte
			w.Header().Set("Content-type", "application/json")

			data, _ = json.Marshal(&v1.ShutdownState{State: state})
			_, _ = w.Write(data)
		}
	}
}

// LogLevelHandler toggles debug logging.
func LogLevelHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	switch r.Method {
	case "POST":
		logLevel := params.ByName("level")
		switch logLevel {
		case LogDefault:
			if <-getLogLevelChan == LogDebug {
				Log.Printf("Log level changed to %s\n", logLevel)
			}
			setLogLevelChan <- logLevel
		case LogDebug:
			if <-getLogLevelChan == LogDefault {
				Log.Printf("Shutdown state changed to %s\n", logLevel)
			}
			setLogLevelChan <- logLevel
		default:
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(400)
			_, _ = w.Write(simpleError(fmt.Errorf("Unsupported log level: %v", logLevel)))
			return
		}
		w.WriteHeader(200)
		return
	}
}

func simpleError(err error) []byte {
	m := v1.Meta{Error: err.Error()}
	data, _ := json.Marshal(&m)
	return data
}
