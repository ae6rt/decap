package deferrals

import (
	"log"
	"sync"
)

// InMemoryDeferralService is the working network deferral service.
type InMemoryDeferralService struct {
	mutex     sync.Mutex
	deferrals []Deferrable

	logger *log.Logger
}

// NewDefault is the constructor for a DeferralService maintained in memory.
func NewDefault(log *log.Logger) DeferralService {
	return &InMemoryDeferralService{logger: log}
}

// Defer puts a build onto the deferred list.
func (t *InMemoryDeferralService) Defer(event Deferrable) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// dedup as we go
	if len(t.deferrals) > 0 && event.Lockname() == t.deferrals[len(t.deferrals)-1].Lockname() {
		return nil
	}

	t.deferrals = append(t.deferrals, event)

	return nil
}

// List lists and dedupes the deferred builds.  Used for presentation in a frontend UI.
func (t *InMemoryDeferralService) List() ([]Deferrable, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.copy(), nil
}

// Poll reads and dedups the list of deferred builds, clears the list from backing store and returns the list to the caller.  Called for
// purposes of relaunching builds.
func (t *InMemoryDeferralService) Poll() ([]Deferrable, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	c := t.copy()

	t.deferrals = nil

	return c, nil
}

// Remove removes a build by ID from the deferred list.
func (t *InMemoryDeferralService) Remove(id string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var idx int
	var found bool

	for j, k := range t.deferrals {
		if k.GetID() == id {
			idx = j
			found = true
			break
		}
	}

	if found {
		t.deferrals = append(t.deferrals[:idx], t.deferrals[idx+1:]...)
	}

	return nil
}

// Do not call this outside a mutex.
func (t *InMemoryDeferralService) copy() []Deferrable {
	c := make([]Deferrable, len(t.deferrals), len(t.deferrals))
	copy(c, t.deferrals)
	return c
}
