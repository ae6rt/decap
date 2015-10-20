package main

var setShutdownChan = make(chan string)
var getShutdownChan = make(chan string)

const (
	BUILD_QUEUE_OPEN  = "open"
	BUILD_QUEUE_CLOSE = "close"
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
