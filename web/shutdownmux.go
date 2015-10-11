package main

type Shutdown string

var setShutdownChan = make(chan Shutdown)
var getShutdownChan = make(chan Shutdown)

const (
	OPEN  Shutdown = "open"
	CLOSE Shutdown = "close"
)

func shutdownMux(initialValue Shutdown) {
	t := initialValue
	Log.Print("Shutdown channel mux running")
	for {
		select {
		case t = <-setShutdownChan:
		case getShutdownChan <- t:
		}
	}
}
