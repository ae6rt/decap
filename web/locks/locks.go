package locks

import (
	"encoding/hex"
	"fmt"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

const BUILDLOCKS = "/buildlocks"
const DEFERRED = "/deferred"

type Locker interface {
	Lock(key, value string) (*etcd.Response, error)
	Unlock(key, value string) (*etcd.Response, error)
	Defer(buildEvent []byte) (*etcd.Response, error)
	ClearDeferred(deferredID string) (*etcd.Response, error)
	DeferredBuilds() ([]Deferral, error)
	InitDeferred() error
	Key(projectKey, branch string) string
}

type Deferral struct {
	Data string
	Key  string
}

type EtcdLocker struct {
	Config etcd.Config
}

func (d EtcdLocker) Lock(key, value string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).Set(context.Background(), BUILDLOCKS+"/"+key, value, &etcd.SetOptions{PrevExist: etcd.PrevNoExist})
}

func (d EtcdLocker) Unlock(key, value string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).Delete(context.Background(), BUILDLOCKS+"/"+key, &etcd.DeleteOptions{PrevValue: value})
}

func (d EtcdLocker) Key(projectKey, branch string) string {
	return hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", projectKey, branch)))
}

func (d EtcdLocker) Defer(buildEvent []byte) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).CreateInOrder(context.Background(), DEFERRED, string(buildEvent), nil)
}

func (d EtcdLocker) ClearDeferred(deferralID string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).Delete(context.Background(), deferralID, nil)
}

func (d EtcdLocker) DeferredBuilds() ([]Deferral, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	resp, err := etcd.NewKeysAPI(c).Get(context.Background(), DEFERRED, &etcd.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		return nil, err
	}

	events := make([]Deferral, 0)
	for _, v := range resp.Node.Nodes {
		d := Deferral{Data: v.Value, Key: v.Key}
		events = append(events, d)
	}
	return events, nil
}

// Create the deferred directory in etcd
func (d EtcdLocker) InitDeferred() error {
	c, err := etcd.New(d.Config)
	if err != nil {
		return err
	}
	_, err = etcd.NewKeysAPI(c).Set(context.Background(), DEFERRED, "", &etcd.SetOptions{Dir: true})
	return err
}

func NewEtcdLocker(machines []string) Locker {
	return EtcdLocker{Config: etcd.Config{
		Endpoints: machines,
		Transport: etcd.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	},
	}
}
