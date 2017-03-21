package v1

import (
	"regexp"

	"github.com/ae6rt/decap/web/locks"
)

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

// UserBuildEvent models an abstract build, independent of the source code management system that backs it.  The fields have a trailing _ because this struct has
// methods with the same name.
type UserBuildEvent struct {
	Team_    string   `json:"team"`
	Project_ string   `json:"project"`
	Refs_    []string `json:"refs"`

	// TODO this feels wrong being here; a build event does not "have" deferrals.   And a Deferral should not be associated with a locker.  Move the defintion of a Deferral to api/v1/types.go (here).  msp 21March2017
	Deferral locks.Deferral `json:"deferral"`
}

// TODO this definition seems strained.  Find out who returns this and who consumes it and probably have those things deal in []Deferral.
// Deferred models a deferred build.  Builds are deferred when they cannot get a lock on their backing branch.  Decap uses locks
// to prevent concurrent builds of the same branch.
type Deferred struct {
	Meta
	DeferredEvents []UserBuildEvent `json:"deferred"`
}

// Deferral models a deferred build.  Any deferral service will deal in this type.
type Deferral struct {
	// Project is the project or team to which the build belongs.
	Project string `json:"project"`

	// Ref is generally the branch to which the build belongs, but could be a git tag.
	Ref string `json:"ref"`

	// ID is a unique id for an instance, and is expected to be the Build ID (likely a UUID).
	ID string `json:"id"`
}
