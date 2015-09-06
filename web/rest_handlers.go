package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

var projectsHandler = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello projects handler")
}

func buildsHandler(storageService StorageService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		project := vars["project"]
		lib := vars["lib"]

		projectKey := fmt.Sprintf("%s/%s", project, lib)

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
		buildList, err = storageService.GetBuildsByProject(Project{projectKey}, since, limit)

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

func buildLogsHandler(storageService StorageService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		buildID := vars["id"]
		var data []byte
		data, _ = storageService.GetConsoleLog(buildID)
		w.Header().Set("Content-type", "application/x-gzip")
		w.Write(data)
	}
}

func buildArtifactsHandler(storageService StorageService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		buildID := vars["id"]
		var data []byte
		data, _ = storageService.GetArtifacts(buildID)
		w.Header().Set("Content-type", "application/x-gzip")
		w.Write(data)
	}
}

var documentRootHandler = func(w http.ResponseWriter, r *http.Request) {
	if data, err := ioutil.ReadFile("index.html"); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v", err))
		w.WriteHeader(500)
	} else {
		w.Header().Set("Content-type", "text/html")
		fmt.Fprint(w, string(data))
	}
}

var versionHandler = func(w http.ResponseWriter, r *http.Request) {
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
