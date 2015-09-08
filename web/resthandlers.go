package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/ae6rt/githubsdk"
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

func ProjectsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		data, err := json.Marshal(&projects)
		if err != nil {
			fmt.Fprintf(w, "%v\n", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, string(data))
	}
}

func ProjectBranchesHandler(githubClientID, githubClientSecret string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		parent := params.ByName("parent")
		library := params.ByName("library")

		project, present := findProject(parent, library)
		if !present {
			w.WriteHeader(404)
			return
		}

		switch project.Descriptor.RepoManager {
		case "github":
			ghClient := githubsdk.NewGithubClient("https://api.github.com", githubClientID, githubClientSecret)
			branches, err := ghClient.GetBranches(project.Parent, project.Library)
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
		project := params.ByName("parent")
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
		buildList, err = storageService.GetBuildsByProject(Project{Parent: project, Library: library}, since, limit)

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

func VersionHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	version := Version{buildInfo}
	data, err := json.Marshal(&version)
	if err != nil {
		fmt.Fprintf(w, "%v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-type", "application/json")
	fmt.Fprint(w, string(data))
}

func HooksHandler(k8s DefaultDecap) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		w.WriteHeader(200)

		repoManager := params.ByName(":repomanager")

		var event PushEvent
		switch repoManager {
		case "github":
			event = GithubEvent{}
		case "stash":
			event = StashEvent{}
		case "bitbucket":
			event = BitBucketEvent{}
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Log.Println(err)
			return
		}
		if err := json.Unmarshal(data, &event); err != nil {
			Log.Println(err)
			return
		}
		Log.Printf("%s hook received: %s\n", repoManager, data)
		go k8s.launchBuild(event)
	}
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if data, err := ioutil.ReadFile("index.html"); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v", err))
		w.WriteHeader(500)
	} else {
		w.Header().Set("Content-type", "text/html")
		fmt.Fprint(w, string(data))
	}
}

func findProject(parent, library string) (Project, bool) {
	for _, v := range getProjects() {
		if v.Parent == parent && v.Library == library {
			return v, true
		}
	}
	return Project{}, false

}

func getProjects() []Project {
	// todo lock with a mutex
	return projects
}
