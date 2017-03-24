package v1

import "regexp"

// Meta models fields common to most publicly exposed types.
type Meta struct {
	Error string `json:"error,omitempty"`
}

// Version models the version of Decap
type Version struct {
	Meta
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	SDK     string `json:"sdk"`
}

// Projects models the collection of projects managed by Decap
type Projects struct {
	Meta
	Projects []Project `json:"projects"`
}

// Project models a specific project.  Sidecars are literal Kubernetes container specs articulated in JSON.  A user can
// include any amount of detail in the side car spec, which will be deserialized into a native Kubernetes container type
// for inclusion in the materialized build pod.
type Project struct {
	Team        string            `json:"team"`
	ProjectName string            `json:"project"`
	Descriptor  ProjectDescriptor `json:"descriptor,omitempty"`
	Sidecars    []string          `json:"sidecars,omitempty"`
}

// ProjectDescriptor models required information about a project.
type ProjectDescriptor struct {
	// Image is the container image the associated project should be built in.
	Image string `json:"buildImage"`
	// RepoManager is the source code management system the project source is contained in.
	RepoManager string `json:"repoManager"`
	// RepoURL is largely informational for the human managing this project.  It is currently unused by Decap.
	RepoURL string `json:"repoUrl"`
	// RepoDescription is a human readable description of this project.
	RepoDescription string `json:"repoDescription"`
	// ManagedRegexStr is a regular expression that defines which refs (branches and tags) is willing to build as a result of
	// a post-commit hook.  Manual builds are not subject to this regex.
	ManagedRefRegexStr string `json:"managedRefRegex"`
	// The formal regex associated with ManagedRefRegexStr above.
	Regex *regexp.Regexp
}

// IsRefManaged is used by Decap to determine if a build should be launched as a result of a post-commit hook on a given ref.
func (d ProjectDescriptor) IsRefManaged(ref string) bool {
	return d.Regex == nil || d.Regex.MatchString(ref)
}

// Builds is a collection of historical Build's.
type Builds struct {
	Meta
	Builds []Build `json:"builds"`
}

// Build models the historical build information available to Decap in the backing storage system.
type Build struct {
	ID         string `json:"id"`
	ProjectKey string `json:"projectKey"`
	Branch     string `json:"branch"`
	Result     int    `json:"result"`
	Duration   uint64 `json:"duration"`
	UnixTime   uint64 `json:"startTime"`
}

// Teams models a collection of logical Team's.
type Teams struct {
	Meta
	Teams []Team `json:"teams"`
}

// Team models the name of a group of developers or a project.
type Team struct {
	Name string `json:"name"`
}

// Refs models a collection of logical Ref's.
type Refs struct {
	Meta
	Refs []Ref `json:"refs"`
}

// Ref models a git branch or tag.
type Ref struct {
	RefID string `json:"ref"`
	Type  string `json:"type"`
}

// ShutdownState models whether the build queue is open or closed.
type ShutdownState struct {
	Meta
	State string `json:"state"`
}

// UserBuildEvent models an abstract build, independent of the source code management system that backs it.
type UserBuildEvent struct {
	Meta

	// Team is the github account owner or BitBucket project key.
	Team_ string `json:"team"`

	// Project is the git repository basename.
	Project_ string `json:"project"`

	// Ref is the branch or tag to be built.
	Ref_ string `json:"ref"`

	// ID is an opaque build ID assigned when a build is scheduled.
	ID string `json:"id"`

	// DeferredUnixTime is the time at which a build is deferred, if at all.
	DeferredUnixtime int64 `json:"deferred-unixtime"`
}
