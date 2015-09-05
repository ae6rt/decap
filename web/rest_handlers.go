package main

import (
	"encoding/json"
	"fmt"
	"github.com/ae6rt/decap/api/v1"
	"io/ioutil"
	"net/http"
)

var projectsHandler = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello projects handler")
}

var buildsHandler = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello build handling")
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
	version := v1.Version{buildInfo}
	data, err := json.Marshal(&version)
	if err != nil {
		fmt.Fprintf(w, "%v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-type", "application/json")
	fmt.Fprint(w, string(data))
}
