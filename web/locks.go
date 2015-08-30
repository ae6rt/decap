package main

import (
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"time"
)

type DefaultLock struct {
	Config etcd.Config
	Locker
}

type Locker interface {
	Lock(a, b string) (*etcd.Response, error)
	Unlock(a, b string) (*etcd.Response, error)
}

func (d DefaultLock) Lock(lockKey, buildID string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	client := etcd.NewKeysAPI(c)
	return client.Set(context.Background(), lockKey, buildID, &etcd.SetOptions{PrevExist: etcd.PrevNoExist})
}

func (d DefaultLock) UnLock(lockKey, buildID string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	client := etcd.NewKeysAPI(c)
	return client.Delete(context.Background(), lockKey, &etcd.DeleteOptions{PrevValue: buildID})
}

func NewDefaultLock(machines []string) DefaultLock {
	return DefaultLock{Config: etcd.Config{
		Endpoints: []string{"http://lockservice:2379"},
		Transport: etcd.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	},
	}
}
