package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestCors(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()

	HandleOptions(w, req, []httprouter.Param{})

	var h string

	h = w.Header().Get("Access-Control-Allow-Origin")
	if h != "http://localhost:9000" {
		t.Fatalf("Want http://localhost:9000 but got %s\n", h)
	}

	h = w.Header().Get("Access-Control-Allow-Headers")
	if h != "DECAP-APP-NAME, DECAP-API-TOKEN, Cache-Control, Pragma, Origin, Authorization, Content-Type, X-Requested-With" {
		t.Fatalf("Want DECAP-APP-NAME, DECAP-API-TOKEN, Cache-Control, Pragma, Origin, Authorization, Content-Type, X-Requested-With but got %s\n", h)
	}

	h = w.Header().Get("Access-Control-Allow-Methods")
	if h != "GET, PUT, POST, DELETE, OPTIONS" {
		t.Fatalf("Want GET, PUT, POST, DELETE, OPTIONS but got %s\n", h)
	}

}
