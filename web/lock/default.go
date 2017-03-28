package lock

import (
	"errors"
	"sync"

	"github.com/ae6rt/decap/web/api/v1"
	"k8s.io/client-go/kubernetes"
	k8sapi "k8s.io/client-go/pkg/api/v1"
)

// DefaultLockService queries the k8s master to find out if a pod is building a project+branch.
type DefaultLockService struct {
	mutex     sync.Mutex
	clientset *kubernetes.Clientset
}

// NewDefaultLockService defines a lock service with an Acquqire method that simply queries
// k8s master for whether a build is running with the input v1.UserBuildEvent's lockname.
func NewDefaultLockService(clientset *kubernetes.Clientset) DistributedLockService {
	return &DefaultLockService{clientset: clientset}
}

// Acquire attempts to acquire a lock on the given object
func (t *DefaultLockService) Acquire(obj v1.UserBuildEvent) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	pods, err := t.clientset.CoreV1().Pods("decap").List(k8sapi.ListOptions{
		LabelSelector: "lockname=" + obj.Lockname(),
	})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return nil
	}
	return errors.New("a build with lockname " + obj.Lockname() + " is already running")
}

// Release is a no op for the default lock service.
func (t *DefaultLockService) Release(obj v1.UserBuildEvent) error {
	return nil
}
