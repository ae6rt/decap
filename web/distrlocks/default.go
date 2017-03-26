package distrlocks

import (
	"errors"
	"sync"

	"github.com/ae6rt/decap/web/api/v1"
	"k8s.io/client-go/kubernetes"
	av1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

// DefaultLockService queries the k8s master to find out if a pod is building a project+branch.
type DefaultLockService struct {
	mutex     sync.Mutex
	clientset *kubernetes.Clientset
}

// NewDefaultLockService defines a lock service with an Acquqire method that simply queries
// k8s master for whether a build is running with the input v1.UserBuildEvent's lockname.
func NewDefaultLockService() (DistributedLockService, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &DefaultLockService{clientset: clientset}, nil

}

// Acquire attempts to acquire a lock on the given object
func (t *DefaultLockService) Acquire(obj v1.UserBuildEvent) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	pods, err := t.clientset.CoreV1().Pods("decap").List(av1.ListOptions{
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
