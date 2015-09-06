package main

type Version struct {
	Version string `json:"version"`
}

type Project struct {
	Key string `json:"key"`
}

type Build struct {
	ID       string `json:"id"`
	Branch   string `json:"branch"`
	Result   int    `json:"result"`
	Duration uint64 `json:"duration"`
	UnixTime uint64 `json:"unixtime"`
}
