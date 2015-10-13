package main

type LogLevel string

var setLogLevelChan = make(chan LogLevel)
var getLogLevelChan = make(chan LogLevel)

const (
	LOG_DEFAULT LogLevel = "default"
	LOG_DEBUG   LogLevel = "debug"
)

func logLevelMux(initialValue LogLevel) {
	t := initialValue
	Log.Print("LogLevel channel mux running")
	for {
		select {
		case t = <-setLogLevelChan:
		case getLogLevelChan <- t:
		}
	}
}
