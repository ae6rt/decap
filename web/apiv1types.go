package main

type Meta struct {
	Error            string `json:"error,omitempty"`
	LastEvaluatedKey string `json:"lastkey,omitempty"`
}

type Version struct {
	Meta
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	SDK     string `json:"sdk"`
}

type Projects struct {
	Meta
	Projects []Project `json:"projects"`
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

type Teams struct {
	Meta
	Teams []Team `json:"teams"`
}

type Team struct {
	Name string `json:"name"`
}
