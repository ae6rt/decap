package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

var documentRootHandler = func(w http.ResponseWriter, r *http.Request) {
	if data, err := ioutil.ReadFile("index.html"); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v", err))
		w.WriteHeader(500)
	} else {
		fmt.Fprint(w, string(data))
	}
}