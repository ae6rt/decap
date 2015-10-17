package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/julienschmidt/httprouter"
)

func TestShutdownBuildQueueOpen(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	getShutdownChan = make(chan v1.Shutdown, 1)
	setShutdownChan = make(chan v1.Shutdown, 1)
	getShutdownChan <- "any incumbent value will do to avoid blocking on the channel read"

	w := httptest.NewRecorder()
	ShutdownHandler(w, req, []httprouter.Param{
		httprouter.Param{Key: "state", Value: string(BUILD_QUEUE_OPEN)},
	},
	)

	if w.Code != 200 {
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}

	state := <-setShutdownChan
	if state != BUILD_QUEUE_OPEN {
		t.Fatalf("Want open but got %s\n", state)
	}
}

func TestShutdownBuildQueueClose(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	getShutdownChan = make(chan v1.Shutdown, 1)
	setShutdownChan = make(chan v1.Shutdown, 1)
	getShutdownChan <- "any incumbent value will do to avoid blocking on the channel read"

	w := httptest.NewRecorder()
	ShutdownHandler(w, req, []httprouter.Param{
		httprouter.Param{Key: "state", Value: string(BUILD_QUEUE_CLOSE)},
	},
	)

	if w.Code != 200 {
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}

	state := <-setShutdownChan
	if state != BUILD_QUEUE_CLOSE {
		t.Fatalf("Want open but got %s\n", state)
	}
}

func TestShutdownBuildQueueInvalid(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	getShutdownChan = make(chan v1.Shutdown, 1)
	setShutdownChan = make(chan v1.Shutdown, 1)
	getShutdownChan <- "any incumbent value will do to avoid blocking on the channel read"

	w := httptest.NewRecorder()
	ShutdownHandler(w, req, []httprouter.Param{
		httprouter.Param{Key: "state", Value: string("nope")},
	},
	)

	if w.Code != 400 {
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
}
