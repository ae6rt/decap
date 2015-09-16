package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

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
	a := make([]Team, 0)
	p := getProjects()
	for _, v := range p {
		a = append(a, Team{Name: v.Team})
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

func ExecuteBuildHandler(decap Decap) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		team := params.ByName("team")
		library := params.ByName("library")

		if _, present := projectByTeamLibrary(team, library); !present {
			w.WriteHeader(404)
			return
		}

		branches := r.URL.Query()["branch"]
		if len(branches) == 0 {
			// todo add a message
			w.WriteHeader(400)
			return
		}

		event := UserBuildEvent{team: team, library: library, branches: branches}
		go decap.LaunchBuild(event)
	}
}

func HooksHandler(buildScriptsRepo, buildScriptsBranch string, decap Decap) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		repoManager := params.ByName("repomanager")
		if !(repoManager == "github" || repoManager == "buildscripts") {
			Log.Printf("repomanager %s not supported\n", repoManager)
			w.WriteHeader(400)
			return
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Log.Println(err)
			w.WriteHeader(500)
			return
		}
		defer func() {
			r.Body.Close()
		}()

		Log.Printf("%s hook received: %s\n", repoManager, data)

		switch repoManager {
		case "buildscripts":
			p, err := assembleProjects(buildScriptsRepo, buildScriptsBranch)
			if err != nil {
				Log.Println(err)
				w.WriteHeader(500)
			} else {
				setProjects(p)
			}
		case "github":
			event := GithubEvent{}
			if err := json.Unmarshal(data, &event); err != nil {
				Log.Println(err)
				w.WriteHeader(500)
				return
			}
			go decap.LaunchBuild(event)
		}

		w.WriteHeader(200)
	}
}

// todo Cleanup with PreStop might help explain to the world the state of the container immediately before termination.
// See https://godoc.org/github.com/kubernetes/kubernetes/pkg/api/v1#Lifecycle
func StopBuildHandler(decap Decap) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		buildID := params.ByName("id")
		if err := decap.DeletePod(buildID); err != nil {
			Log.Println(err)
		}
	}
}

func ProjectBranchesHandler(repoClients map[string]SCMClient) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		team := params.ByName("team")
		library := params.ByName("library")

		project, present := projectByTeamLibrary(team, library)
		if !present {
			w.WriteHeader(404)
			return
		}

		switch project.Descriptor.RepoManager {
		case "github":
			w.Header().Set("Content-type", "application/json")
			repoClient := repoClients["github"]
			nativeBranches, err := repoClient.GetBranches(project.Team, project.Library)
			if err != nil {
				Log.Print(err)
				data, _ := json.Marshal(&Branches{Meta: Meta{Error: err.Error()}})
				w.WriteHeader(500)
				w.Write(data)
				return
			}

			branches := Branches{Branches: nativeBranches}
			data, err := json.Marshal(&branches)
			if err != nil {
				Log.Print(err)
				data, _ := json.Marshal(&Branches{Meta: Meta{Error: err.Error()}})
				w.WriteHeader(500)
				w.Write(data)
				return
			}

			w.Write(data)
			return
		}
		w.WriteHeader(400)
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
			return
		}
		if r.Header.Get("Accept") == "text/plain" {
			inputBuffer := bytes.NewBuffer(data)
			source, err := gzip.NewReader(inputBuffer)
			if err != nil {
				Log.Println(err)
				w.WriteHeader(500)
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
			return
		}
		if r.Header.Get("Accept") == "text/plain" {
			// unzip
			inputBuffer := bytes.NewBuffer(data)
			source, err := gzip.NewReader(inputBuffer)
			if err != nil {
				Log.Println(err)
				w.WriteHeader(500)
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
		project := params.ByName("team")
		library := params.ByName("library")

		since, err := toUint64(r.URL.Query().Get("since"), 0)
		if err != nil {
			builds := Builds{Meta: Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(400)
			w.Write(data)
			return
		}

		limit, err := toUint64(r.URL.Query().Get("limit"), math.MaxUint64)
		if err != nil {
			builds := Builds{Meta: Meta{Error: err.Error()}}
			data, _ := json.Marshal(&builds)
			w.WriteHeader(400)
			w.Write(data)
			return
		}

		buildList, err := storageService.GetBuildsByProject(Project{Team: project, Library: library}, since, limit)

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
