package main

import (
	"encoding/json"
	"fmt"
	"github.com/ae6rt/decap/api/v1"
	"net/http"
)

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
