package main

var setLogLevelChan = make(chan string)
var getLogLevelChan = make(chan string)

const (
	// The verb POSTed to set logging to the default level.
	LogDefault = "default"
	// The verb POSTed to set logging to debug level.
	LogDebug = "debug"
)

func logLevelMux(initialValue string) {
	t := initialValue
	Log.Print("LogLevel channel mux running")
	for {
		select {
		case t = <-setLogLevelChan:
		case getLogLevelChan <- t:
		}
	}
}
