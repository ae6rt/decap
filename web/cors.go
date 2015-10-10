package main

import (
	"net/http"
)

var corsWrapper = func(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeCorsHeaders(w)
		handler.ServeHTTP(w, r)
	})
}

func writeCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:9000")
	w.Header().Set("Access-Control-Allow-Headers", "DECAP-APP-NAME, DECAP-API-TOKEN, Cache-Control, Pragma, Origin, Authorization, Content-Type, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
}
