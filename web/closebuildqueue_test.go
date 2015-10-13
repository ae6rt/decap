package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestOpenBuildQueue(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	w := httptest.NewRecorder()

	getShutdownChan = make(chan Shutdown, 1)
	setShutdownChan = make(chan Shutdown, 1)

	// any incumbent value will do to avoid blocking on the channel read
	getShutdownChan <- "anyvalue"

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

func TestCloseBuildQueue(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	w := httptest.NewRecorder()

	getShutdownChan = make(chan Shutdown, 1)
	setShutdownChan = make(chan Shutdown, 1)

	// any incumbent value will do to avoid blocking on the channel read
	getShutdownChan <- "anyvalue"

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

func TestInvalidShutdownBuildQueue(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	w := httptest.NewRecorder()

	getShutdownChan = make(chan Shutdown, 1)
	setShutdownChan = make(chan Shutdown, 1)

	// any incumbent value will do to avoid blocking on the channel read
	getShutdownChan <- "anyvalue"

	ShutdownHandler(w, req, []httprouter.Param{
		httprouter.Param{Key: "state", Value: string("nope")},
	},
	)

	if w.Code != 400 {
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
}
