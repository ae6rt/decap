package deferrals

// DeferralService models how builds are deferred and rehydrated for execution.
type DeferralService interface {
	Defer(projectKey, branch string) error
	Resubmit()
}

// Deferral models a deferred build
type Deferral struct {
	ProjectKey string
	Branch     string
	UnixTime   int64
}
