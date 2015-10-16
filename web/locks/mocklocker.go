package locks

import etcd "github.com/coreos/etcd/client"

type NoOpLocker struct {
}

func (noop NoOpLocker) Lock(key, value string) (*etcd.Response, error) {
	return nil, nil
}

func (noop NoOpLocker) Unlock(key, value string) (*etcd.Response, error) {
	return nil, nil
}

func (noop NoOpLocker) Defer(data []byte) (*etcd.Response, error) {
	return nil, nil
}

func (noop NoOpLocker) ClearDeferred(deferralID string) (*etcd.Response, error) {
	return nil, nil
}

func (noop NoOpLocker) DeferredBuilds() ([]Deferral, error) {
	return nil, nil
}

func (noop NoOpLocker) InitDeferred() error {
	return nil
}

func (noop NoOpLocker) Key(projectKey, branch string) string {
	return "opaquekey"
}
