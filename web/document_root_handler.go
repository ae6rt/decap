package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var documentRootHandler = func(w http.ResponseWriter, r *http.Request) {
	if data, err := ioutil.ReadFile("index.html"); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v", err))
		w.WriteHeader(500)
	} else {
		w.Header().Set("Content-type", "text/html")
		fmt.Fprint(w, string(data))
	}
}
