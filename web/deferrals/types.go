package deferrals

// DeferralService models how builds are deferred and rehydrated for execution.
type DeferralService interface {
	Defer(projectKey, branch, buildID string) error
	Resubmit()
}

// Deferral models a deferred build
type Deferral struct {
	ProjectKey string
	Branch     string
	BuildID    string
	UnixTime   int64
}
