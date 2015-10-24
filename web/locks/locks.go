package locks

import (
	"encoding/hex"
	"fmt"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

const buildLocks = "/buildlocks"
const deferred = "/deferred"

// Locker defines the interface the build locker and build deferral instance makes available.
type Locker interface {
	// Lock sets a lock in the lock service to prevent concurrent builds of a given branch
	Lock(key, value string) (*etcd.Response, error)
	// Unlock clears a lock on a build
	Unlock(key, value string) (*etcd.Response, error)
	// Defer creates an entry in the lock service that marks a build as deferred.
	Defer(buildEvent []byte) (*etcd.Response, error)
	// Clears the entry for a deferred build
	ClearDeferred(deferredID string) (*etcd.Response, error)
	// Returns the current list of deferred builds in created order
	DeferredBuilds() ([]Deferral, error)
	// Initialize the deferred build system if necessary
	InitDeferred() error
	// Form the opaque key for a given branch on a project
	Key(projectKey, branch string) string
}

// Deferral models a deferred build.  Data is the serialized build event, while key is the identifier assigned to this deferred
// build by the locker service.
type Deferral struct {
	// The underlying opaque build event.
	Data string `json:"-"`
	// The key under which the deferred build is stored.
	Key string `json:"key"`
}

// EtcdLocker is the Locker implementation on top of etcd.
type EtcdLocker struct {
	// Config is the underlying etcd cluster configuration
	Config etcd.Config
}

// Lock locks a build on top of etcd.
func (d EtcdLocker) Lock(key, value string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).Set(context.Background(), buildLocks+"/"+key, value, &etcd.SetOptions{PrevExist: etcd.PrevNoExist})
}

// Unlock unlocks a build on top of etcd.
func (d EtcdLocker) Unlock(key, value string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).Delete(context.Background(), buildLocks+"/"+key, &etcd.DeleteOptions{PrevValue: value})
}

// Key defines an unique, opaque key that locks a build.
func (d EtcdLocker) Key(projectKey, branch string) string {
	return hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", projectKey, branch)))
}

// Defer defers a build by saving it to a locker service provided key.  See Atomically Creating In-Order Keys
// at https://coreos.com/etcd/docs/0.4.7/etcd-api/.
func (d EtcdLocker) Defer(buildEvent []byte) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).CreateInOrder(context.Background(), deferred, string(buildEvent), nil)
}

// ClearDeferred removes a build from the deferred list.
func (d EtcdLocker) ClearDeferred(deferralID string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).Delete(context.Background(), deferralID, nil)
}

// DeferredBuilds returns the list of builds currently deferred.
func (d EtcdLocker) DeferredBuilds() ([]Deferral, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	resp, err := etcd.NewKeysAPI(c).Get(context.Background(), deferred, &etcd.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		return nil, err
	}

	var events []Deferral
	for _, v := range resp.Node.Nodes {
		d := Deferral{Data: v.Value, Key: v.Key}
		events = append(events, d)
	}
	return events, nil
}

// InitDeferred creates the directory in etcd where deferred builds are stored.
func (d EtcdLocker) InitDeferred() error {
	_, err := d.DeferredBuilds()
	if err != nil {
		var c etcd.Client
		c, err = etcd.New(d.Config)
		if err != nil {
			return err
		}
		_, err = etcd.NewKeysAPI(c).Set(context.Background(), deferred, "", &etcd.SetOptions{Dir: true})
	}
	return err
}

// New returns a new etcd implementation of a Locker.
func NewEtcdLocker(machines []string) Locker {
	return EtcdLocker{Config: etcd.Config{
		Endpoints: machines,
		Transport: etcd.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	},
	}
}
