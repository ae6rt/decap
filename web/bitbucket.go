package main

import (
	"io/ioutil"
	"net/http"
)

type BitBucketEvent struct {
	PushEvent
}

func (event BitBucketEvent) ProjectKey() string {
	return ""
}

func (event BitBucketEvent) Branches() []string {
	return []string{}
}

type BitBucketHandler struct {
	K8sBase K8sBase
	Handler
}

func (handler BitBucketHandler) handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Println(err)
		return
	}
	Log.Printf("BitBucket hook received: %s\n", data)

	go handler.K8sBase.launchBuild(BitBucketEvent{})
}
