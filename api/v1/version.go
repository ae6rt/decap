package v1

type Version struct {
	Version string `json:"version"`
}

type ProjectList struct {
	Projects []Project `json:"projects"`
}

type Project struct {
	Key string `json:"key"`
}

type Build struct {
	ID       string `json:"id"`
	Branch   string `json:"branch"`
	Duration int64  `json:"duration"`
	Result   int    `json:"result"`
	UnixTime int64  `json:"unixtime"`
}
