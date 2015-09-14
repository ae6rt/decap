package main

import (
	"encoding/json"
	"fmt"
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
	data, err := json.Marshal(&version)
	if err != nil {
		fmt.Fprintf(w, "%v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-type", "application/json")
	fmt.Fprint(w, string(data))
}

func TeamsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a := make([]Team, 0)
	p := getProjects()
	for _, v := range p {
		a = append(a, Team{Name: v.Team})
	}
	teams := Teams{Teams: a}
	data, err := json.Marshal(&teams)
	if err != nil {
		fmt.Fprintf(w, "%v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-type", "application/json")
	fmt.Fprint(w, string(data))
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

	p := Projects{Projects: arr}
	data, err := json.Marshal(&p)
	if err != nil {
		fmt.Fprintf(w, "%v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-type", "application/json")
	fmt.Fprint(w, string(data))
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
			repoClient := repoClients["github"]
			branches, err := repoClient.GetBranches(project.Team, project.Library)
			if err != nil {
				// todo put error on json object
				Log.Print(err)
				w.WriteHeader(500)
				fmt.Fprintf(w, "%v\n", err)
				return
			}
			data, err := json.Marshal(&branches)
			if err != nil {
				w.WriteHeader(500)
				// todo put error on json object
				fmt.Fprintf(w, "%v\n", err)
				return
			}
			w.Write(data)
			return
		}
		w.WriteHeader(404)
	}
}

func LogHandler(storageService StorageService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		buildID := params.ByName("id")
		var data []byte
		data, _ = storageService.GetConsoleLog(buildID)
		w.Header().Set("Content-type", "application/x-gzip")
		w.Write(data)
	}
}

func ArtifactsHandler(storageService StorageService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		buildID := params.ByName("id")
		var data []byte
		data, _ = storageService.GetArtifacts(buildID)
		w.Header().Set("Content-type", "application/x-gzip")
		w.Write(data)
	}
}

func BuildsHandler(storageService StorageService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		project := params.ByName("team")
		library := params.ByName("library")

		var builds Builds

		since, err := toUint64(r.URL.Query().Get("since"), 0)
		if err != nil {
			builds.Meta.Error = fmt.Sprintf("%v", err)
			var data []byte
			data, _ = json.Marshal(&builds)
			fmt.Fprintf(w, "%s", string(data))
			return
		}

		limit, err := toUint64(r.URL.Query().Get("limit"), math.MaxUint64)
		if err != nil {
			builds.Meta.Error = fmt.Sprintf("%v", err)
			var data []byte
			data, _ = json.Marshal(&builds)
			fmt.Fprintf(w, "%s", string(data))
			return
		}

		var buildList []Build
		buildList, err = storageService.GetBuildsByProject(Project{Team: project, Library: library}, since, limit)

		if err != nil {
			builds.Meta.Error = fmt.Sprintf("%v", err)
			var data []byte
			data, _ = json.Marshal(&builds)
			fmt.Fprintf(w, "%s", string(data))
			return
		}

		builds.Builds = buildList
		data, err := json.Marshal(&builds)
		if err != nil {
			builds.Meta.Error = fmt.Sprintf("%v", err)
			var data []byte
			data, _ = json.Marshal(&builds)
			fmt.Fprintf(w, "%s", string(data))
			return
		}
		fmt.Fprint(w, string(data))
	}
}
