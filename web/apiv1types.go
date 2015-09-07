package main

type Meta struct {
	Error            string `json:"error,omitempty"`
	LastEvaluatedKey string `json:"lastkey,omitempty"`
}

type Version struct {
	Version string `json:"version"`
}

type Projects struct {
	Meta
	Projects []Projects `json:"projects"`
}

type Project struct {
	Parent     string            `json:"key"`
	Library    string            `json:"library"`
	Descriptor ProjectDescriptor `json:"descriptor"`
}

type ProjectDescriptor struct {
	RepoManager     string `json:"repo-manager"`
	RepoURL         string `json:"repo-url"`
	RepoDescription string `json:"repo-description"`
}

type Builds struct {
	Meta
	Builds []Build `json:"builds"`
}

type Build struct {
	ID       string `json:"id"`
	Branch   string `json:"branch"`
	Result   int    `json:"result"`
	Duration uint64 `json:"duration"`
	UnixTime uint64 `json:"unixtime"`
}
