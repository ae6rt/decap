package deferrals

/*
A build is defined by a v1.UserBuildEvent type.  This type has a team, a project, a branch, and a unique runtime-assigned UUID.
For Github codebases, the team is the Github account owner, the project is the repository basename, and the branch is the branch to
build.

A build is deferred when it cannot be successfully scheduled on the build cluster.  When this happens,
the build event is timestamped and put on a list of builds to run later.

Builds are periodically read from the deferred list and relaunched in the order that they are timestamped.  If two time-consecutive builds in the deferred list have
the same project+branch, these records are deduped and considered one deferred build.  A build is removed from the deferred
build list the moment it is submitted for relaunch.  If that relaunch also fails, the build is put back onto the deferred list.

Builds can be manually removed from the deferred list by calling Remove() with the buildID.  Used by the frontend UI so the operator
can manually remove a build from the deferred list.
*/

type Deferrable interface {
	GetID() string
	Lockname() string
}

// DeferralService models how builds are deferred and relaunched later.
type DeferralService interface {
	// Defer puts a build onto the deferred list.
	Defer(Deferrable) error

	// List lists and dedupes the deferred builds.  Used for presentation in a frontend UI.
	List() ([]Deferrable, error)

	// Poll reads and dedups the list of deferred builds, clears the list from backing store and returns the list to the caller.  Called for
	// purposes of relaunching builds.
	Poll() ([]Deferrable, error)

	// Remove removes a build by ID from the deferred list.
	Remove(id string) error
}
