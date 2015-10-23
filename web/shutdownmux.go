package main

var setShutdownChan = make(chan string)
var getShutdownChan = make(chan string)

const (
	// The verb POSTed to open the build queue
	BuildQueueOpen = "open"
	// The verb POSTed to close the build queue
	BuildQueueClose = "close"
)

func shutdownMux(initialValue string) {
	t := initialValue
	Log.Print("Shutdown channel mux running")
	for {
		select {
		case t = <-setShutdownChan:
		case getShutdownChan <- t:
		}
	}
}
