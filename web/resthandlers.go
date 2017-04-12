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

// TeamsHandler returns information about managed teams.
func TeamsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	p := getProjects()

	keys := make(map[string]string)
	for _, v := range p {
		keys[v.Team] = ""
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

// ProjectsHandler returns informtion about managed projects.
func ProjectsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	team := r.URL.Query().Get("team")
	var arr []v1.Project
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
		_, _ = w.Write(data)
		return
	}

	_, _ = w.Write(data)
}

// DeferredBuildsHandler returns information about deferred builds.
func DeferredBuildsHandler(builder Builder) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		switch r.Method {
		case "GET":
			deferred, err := builder.DeferredBuilds()
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
			w.Header().Set("Content-type", "application/json")
			_, _ = w.Write(data)
		case "POST":
			key := r.URL.Query().Get("key")
			if key == "" {
				w.WriteHeader(400)
				_, _ = w.Write(simpleError(fmt.Errorf("Missing or empty key parameter in clear deferred build")))
				return
			}
			if err := builder.ClearDeferredBuild(key); err != nil {
				w.WriteHeader(500)
				data, _ := json.Marshal(&v1.UserBuildEvent{Meta: v1.Meta{Error: err.Error()}})
				_, _ = w.Write(data)
			}
		}
	}
}

// ExecuteBuildHandler handles user-requested build executions.
func ExecuteBuildHandler(decap Builder) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		team := params.ByName("team")
		project := params.ByName("project")

		if _, present := projectByTeamName(team, project); !present {
			Log.Printf("Unknown project %s/%s", team, project)
			w.WriteHeader(404)
			_, _ = w.Write(simpleError(fmt.Errorf("Unknown project %s/%s", team, project)))
			return
		}

		branches := r.URL.Query()["branch"]
		if len(branches) == 0 {
			Log.Println("No branches specified")
			w.WriteHeader(400)
			_, _ = w.Write(simpleError(fmt.Errorf("No branches specified")))
			return
		}

		for _, b := range branches {
			event := v1.UserBuildEvent{Team: team, Project: project, Ref: b}
			go func() {
				_ = decap.LaunchBuild(event)
			}()
		}
	}
}

// HooksHandler handles externally originated SCM events that trigger builds or build-scripts repository refreshes.
func HooksHandler(buildScripts BuildScripts, decap Builder) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		repoManager := params.ByName("repomanager")

		if r.Body == nil {
			Log.Println("Expecting an HTTP entity")
			w.WriteHeader(400)
			return
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Log.Printf("Error reading hook payload: %v\n", err)
			w.WriteHeader(400)
			return
		}
		defer func() {
			_ = r.Body.Close()
		}()

		switch repoManager {
		case "buildscripts":
			if p, err := assembleProjects(buildScripts); err != nil {
				Log.Printf("Error refreshing build scripts: %v\n", err)
				w.WriteHeader(500)
				return
			} else {
				setProjects(p)
				Log.Println("Build scripts refreshed.")
			}
		case "github":
			var event GithubEvent

			eventType := r.Header.Get("X-Github-Event")
			switch eventType {
			case "create":
			case "push":
			default:
				Log.Printf("Unhandled Github event type <%s>.  See https://developer.github.com/webhooks/#delivery-headers.", eventType)
				w.WriteHeader(400)
				return
			}

			if err := json.Unmarshal(data, &event); err != nil {
				Log.Printf("Error unmarshaling Github event: %v\n", err)
				w.WriteHeader(500)
				return
			}

			go func() {
				if err := decap.LaunchBuild(event.BuildEvent()); err != nil {
					Log.Printf("Error launching build for event %+v: %v\n", event, err)
					w.WriteHeader(500)
					return
				}
			}()
		default:
			Log.Printf("repomanager %s not supported\n", repoManager)
			w.WriteHeader(400)
			return
		}

		w.WriteHeader(200)
	}
}

// StopBuildHandler deletes the pod executing the specified build ID.
func StopBuildHandler(decap Builder) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		buildID := params.ByName("id")
		if err := decap.DeletePod(buildID); err != nil {
			Log.Println(err)
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(500)
			_, _ = w.Write(simpleError(err))
		}
	}
}

// ProjectRefsHandler handles informational requests for branches and tags on a project
func ProjectRefsHandler(repoClients map[string]scmclients.SCMClient) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		team := params.ByName("team")
		projectName := params.ByName("project")

		project, present := projectByTeamName(team, projectName)
		if !present {
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(404)
			_, _ = w.Write(simpleError(fmt.Errorf("Unknown project %s/%s", team, projectName)))
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
				_, _ = w.Write(data)
				return
			}

			branches := v1.Refs{Refs: nativeBranches}
			data, err := json.Marshal(&branches)
			if err != nil {
				Log.Print(err)
				data, _ := json.Marshal(&v1.Refs{Meta: v1.Meta{Error: err.Error()}})
				w.WriteHeader(500)
				_, _ = w.Write(data)
				return
			}

			_, _ = w.Write(data)
			return
		default:
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(400)
			_, _ = w.Write(simpleError(fmt.Errorf("repomanager not supported: %s", repositoryManager)))
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

// ShutdownHandler stops the build queue from accepting new build requests.
func ShutdownHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	switch r.Method {
	case "POST":
		shutdownState := params.ByName("state")
		switch shutdownState {
		case BuildQueueClose:
			if <-getShutdownChan == BuildQueueOpen {
				Log.Printf("Shutdown state changed to %s\n", shutdownState)
			}
			setShutdownChan <- shutdownState
		case BuildQueueOpen:
			if <-getShutdownChan == BuildQueueClose {
				Log.Printf("Shutdown state changed to %s\n", shutdownState)
			}
			setShutdownChan <- shutdownState
		default:
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(400)
			_, _ = w.Write(simpleError(fmt.Errorf("Unsupported shutdown state: %v", shutdownState)))
			return
		}
	case "GET":
		var data []byte
		w.Header().Set("Content-type", "application/json")
		data, _ = json.Marshal(&v1.ShutdownState{State: <-getShutdownChan})
		_, _ = w.Write(data)
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
