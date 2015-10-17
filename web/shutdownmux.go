package main

import "github.com/ae6rt/decap/web/api/v1"

var setShutdownChan = make(chan v1.Shutdown)
var getShutdownChan = make(chan v1.Shutdown)

const (
	BUILD_QUEUE_OPEN  v1.Shutdown = "open"
	BUILD_QUEUE_CLOSE v1.Shutdown = "close"
)

func shutdownMux(initialValue v1.Shutdown) {
	t := initialValue
	Log.Print("Shutdown channel mux running")
	for {
		select {
		case t = <-setShutdownChan:
		case getShutdownChan <- t:
		}
	}
}
