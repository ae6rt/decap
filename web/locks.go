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
	Lock(key, value string) (*etcd.Response, error)
	Unlock(key, value string) (*etcd.Response, error)
}

func (d DefaultLock) Lock(key, value string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	client := etcd.NewKeysAPI(c)
	return client.Set(context.Background(), key, value, &etcd.SetOptions{PrevExist: etcd.PrevNoExist})
}

func (d DefaultLock) UnLock(key, value string) (*etcd.Response, error) {
	c, err := etcd.New(d.Config)
	if err != nil {
		return nil, err
	}
	client := etcd.NewKeysAPI(c)
	return client.Delete(context.Background(), key, &etcd.DeleteOptions{PrevValue: value})
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
