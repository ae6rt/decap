package main

type Meta struct {
	Error string `json:"error,omitempty"`
}

type Version struct {
	Meta
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	SDK     string `json:"sdk"`
}

type Atoms struct {
	Meta
	Atoms []Atom `json:"projects"`
}

type Atom struct {
	Team       string         `json:"team"`
	Library    string         `json:"project"`
	Descriptor AtomDescriptor `json:"descriptor,omitempty"`
	Sidecars   []string       `json:"sidecars,omitempty"`
}

type AtomDescriptor struct {
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
	ID         string `json:"id"`
	ProjectKey string `json:"project-key"`
	Branch     string `json:"branch"`
	Result     int    `json:"result"`
	Duration   uint64 `json:"duration"`
	UnixTime   uint64 `json:"start-time"`
}

type Teams struct {
	Meta
	Teams []Team `json:"teams"`
}

type Team struct {
	Name string `json:"name"`
}

type Ref struct {
	RefID string `json:"ref"`
	Type  string `json:"type"`
}

type Refs struct {
	Meta
	Refs []Ref `json:"refs"`
}
