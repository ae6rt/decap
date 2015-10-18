package locks

import etcd "github.com/coreos/etcd/client"

type NoOpLocker struct {
	Data  []byte
	Error error
}

func (noop *NoOpLocker) Lock(key, value string) (*etcd.Response, error) {
	return nil, noop.Error
}

func (noop *NoOpLocker) Unlock(key, value string) (*etcd.Response, error) {
	return nil, noop.Error
}

func (noop *NoOpLocker) Defer(data []byte) (*etcd.Response, error) {
	noop.Data = data
	//	fmt.Printf("internal: %+v\n", noop)
	return nil, noop.Error
}

func (noop *NoOpLocker) ClearDeferred(deferralID string) (*etcd.Response, error) {
	return nil, noop.Error
}

func (noop *NoOpLocker) DeferredBuilds() ([]Deferral, error) {
	return nil, noop.Error
}

func (noop *NoOpLocker) InitDeferred() error {
	return noop.Error
}

func (noop *NoOpLocker) Key(projectKey, branch string) string {
	return "opaquekey"
}
