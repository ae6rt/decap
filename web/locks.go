package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type DefaultLock struct {
	Config etcd.Config
}

type Locker interface {
	Lock(key, value string) (*etcd.Response, error)
	Unlock(key, value string) (*etcd.Response, error)
	Defer(buildEvent []byte) (*etcd.Response, error)
	ClearDeferred(deferredID string) (*etcd.Response, error)
	DeferredBuilds() ([]UserBuildEvent, error)
	Key(projectKey, branch string) string
}

func (d DefaultLock) Lock(key, value string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).Set(context.Background(), "/buildlocks/"+key, value, &etcd.SetOptions{PrevExist: etcd.PrevNoExist})
}

func (d DefaultLock) Unlock(key, value string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	return etcd.NewKeysAPI(c).Delete(context.Background(), "/buildlocks/"+key, &etcd.DeleteOptions{PrevValue: value})
}

func (d DefaultLock) Key(projectKey, branch string) string {
	return hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", projectKey, branch)))
}

func (d DefaultLock) Defer(buildEvent []byte) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}

	// See Atomically Creating In-Order Keys at https://coreos.com/etcd/docs/0.4.7/etcd-api/
	return etcd.NewKeysAPI(c).CreateInOrder(context.Background(), "/deferred", string(buildEvent), nil)
}

func (d DefaultLock) ClearDeferred(deferredID string) (*etcd.Response, error) {
	return nil, nil
}

func (d DefaultLock) DeferredBuilds() ([]UserBuildEvent, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	resp, err := etcd.NewKeysAPI(c).Get(context.Background(), "/deferred", &etcd.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	// See Atomically Creating In-Order Keys at https://coreos.com/etcd/docs/0.4.7/etcd-api/
	events := make([]UserBuildEvent, 0)
	for _, v := range resp.Node.Nodes {
		var o UserBuildEvent
		err := json.Unmarshal([]byte(v.Value), &o)
		if err != nil {
			log.Printf("Error deserializing build event %s: %v\n", v.Key, err)
			continue
		}
		events = append(events, o)
	}
	return events, nil
}

func NewDefaultLock(machines []string) DefaultLock {
	return DefaultLock{Config: etcd.Config{
		Endpoints: machines,
		Transport: etcd.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	},
	}
}
