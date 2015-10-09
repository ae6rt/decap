package main

import (
	"encoding/hex"
	"fmt"
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
	return etcd.NewKeysAPI(c).CreateInOrder(context.Background(), "/deferred", string(buildEvent), nil)
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
