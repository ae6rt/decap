package main

var setLogLevelChan = make(chan string)
var getLogLevelChan = make(chan string)

const (
	LOG_DEFAULT = "default"
	LOG_DEBUG   = "debug"
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
