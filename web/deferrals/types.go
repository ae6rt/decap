package deferrals

import "github.com/ae6rt/decap/web/api/v1"

// DeferralService models how builds are deferred and rehydrated for execution.
type DeferralService interface {
	Defer(v1.UserBuildEvent) error
	List() ([]v1.UserBuildEvent, error)
	Remove(teamProjectKey string) error
	Resubmit()
}
