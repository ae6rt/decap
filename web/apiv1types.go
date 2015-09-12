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
	Team       string            `json:"team"`
	Library    string            `json:"library"`
	Descriptor ProjectDescriptor `json:"descriptor,omitempty"`
	Sidecars   []string          `json:"sidecars,omitempty"`
}

type ProjectDescriptor struct {
	Image           string `json:"build-image"`
	RepoManager     string `json:"repo-manager"`
	RepoURL         string `json:"repo-url"`
	RepoDescription string `json:"repo-description"`
}

type Sidecar string

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
