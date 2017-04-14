package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

/*
type MockShutdown struct {
	BuildManagerBaseMock
	open  bool
	close bool
}

func (t *MockShutdown) CloseQueue() {
	t.close = true
}

func (t *MockShutdown) OpenQueue() {
	t.open = true
}
*/

func TestShutdownBuildQueue(t *testing.T) {
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
	}

	for testNumber, test := range tests {
		req, _ := http.NewRequest("POST", "http://example.com", nil)

		getShutdownChan = make(chan string, 1)
		setShutdownChan = make(chan string, 1)
		getShutdownChan <- "any incumbent value will do to avoid blocking on the channel read"

		buildManager := &DefaultBuildManager{}
		handler := ShutIt(buildManager)

		w := httptest.NewRecorder()

		handler(w, req, []httprouter.Param{httprouter.Param{Key: "state", Value: test.desiredState}})

		if w.Code != test.wantHTTPResponseCode {
			t.Errorf("Test %d: want %d, got %d\n", testNumber, test.wantHTTPResponseCode, w.Code)
		}

		switch test.desiredState {
		case BuildQueueClose:
			if buildManager.QueueIsOpen() {
				t.Errorf("Test %d: expecting build queue to be closed\n", testNumber)
			}
		case BuildQueueOpen:
			if buildManager.QueueIsOpen() {
				t.Errorf("Test %d: expecting build queue to be closed\n", testNumber)
			}
		default:

		}

		/*
			state := <-setShutdownChan
			if state != BuildQueueOpen {
				t.Fatalf("Want open but got %s\n", state)
			}
		*/

	}

}

/*
func TestShutdownBuildQueueClose(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	getShutdownChan = make(chan string, 1)
	setShutdownChan = make(chan string, 1)
	getShutdownChan <- "any incumbent value will do to avoid blocking on the channel read"

	w := httptest.NewRecorder()
	ShutIt(w, req, []httprouter.Param{
		httprouter.Param{Key: "state", Value: BuildQueueClose},
	},
	)

	if w.Code != 200 {
		t.Fatalf("Want 200 but got %d\n", w.Code)
	}

	state := <-setShutdownChan
	if state != BuildQueueClose {
		t.Fatalf("Want open but got %s\n", state)
	}
}

func TestShutdownBuildQueueInvalid(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	getShutdownChan = make(chan string, 1)
	setShutdownChan = make(chan string, 1)
	getShutdownChan <- "any incumbent value will do to avoid blocking on the channel read"

	w := httptest.NewRecorder()
	ShutIt(w, req, []httprouter.Param{
		httprouter.Param{Key: "state", Value: string("nope")},
	},
	)

	if w.Code != 400 {
		t.Fatalf("Want 400 but got %d\n", w.Code)
	}
}
*/
