package main

import (
	"fmt"
	"net/http"
)

var projectsAPIHandler = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello api")
}
