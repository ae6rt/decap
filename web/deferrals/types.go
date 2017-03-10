package deferrals

// DeferralService models how builds are deferred and rehydrated for execution.
type DeferralService interface {
	CreateQueue(queueName string) error
	Defer(projectKey, branch string) error
	Resubmit()
}

// Deferral models a deferred build
type Deferral struct {
	ProjectKey string
	Branch     string
	UnixTime   int64 // there is currently no use for this, but there might be in the future
}
