package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestStopBuildHandler(t *testing.T) {
	req, err := http.NewRequest("DELETE", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	decap := MockDecap{}

	w := httptest.NewRecorder()
	StopBuildHandler(&decap)(w, req, []httprouter.Param{
		httprouter.Param{Key: "id", Value: "the-build-id"},
	},
	)

	if decap.buildID != "the-build-id" {
		log.Fatalf("Want the-build-id but got %s\n", decap.buildID)
	}
}
