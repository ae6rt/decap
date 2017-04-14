package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestToggleBuildQueue(t *testing.T) {
	var tests = []struct {
		desiredState         string
		wantHTTPResponseCode int
	}{
		{
			desiredState:         BuildQueueOpen,
			wantHTTPResponseCode: 200,
		},
		{
			desiredState:         BuildQueueClose,
			wantHTTPResponseCode: 200,
		},
		{
			desiredState:         "<unsupported>",
			wantHTTPResponseCode: 400,
		},
	}

	for testNumber, test := range tests {
		req, _ := http.NewRequest("POST", "http://example.com", nil)

		getShutdownChan = make(chan string, 1)
		setShutdownChan = make(chan string, 1)
		getShutdownChan <- "any incumbent value will do to avoid blocking on the channel read"

		buildManager := &DefaultBuildManager{logger: log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)}
		handler := ShutIt(buildManager)

		w := httptest.NewRecorder()

		handler(w, req, []httprouter.Param{httprouter.Param{Key: "state", Value: test.desiredState}})

		switch test.desiredState {
		case BuildQueueClose:
			if w.Code != test.wantHTTPResponseCode {
				t.Errorf("Test %d: want %d, got %d\n", testNumber, test.wantHTTPResponseCode, w.Code)
			}
			if buildManager.QueueIsOpen() {
				t.Errorf("Test %d: expecting build queue to be closed\n", testNumber)
			}
		case BuildQueueOpen:
			if w.Code != test.wantHTTPResponseCode {
				t.Errorf("Test %d: want %d, got %d\n", testNumber, test.wantHTTPResponseCode, w.Code)
			}
			if buildManager.QueueIsOpen() {
				t.Errorf("Test %d: expecting build queue to be closed\n", testNumber)
			}
		default:
			if w.Code != test.wantHTTPResponseCode {
				t.Errorf("Test %d: want %d, got %d\n", testNumber, test.wantHTTPResponseCode, w.Code)
			}
		}
	}
}
