package distrlocks

import (
	"fmt"

	"github.com/ae6rt/decap/web/api/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// DefaultLockService queries the k8s master to find out if a pod is building a project+branch.
type DefaultLockService struct {
	clientset *kubernetes.Clientset
}

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
	fmt.Println(t.clientset)
	/*
		pods, err := t.clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	*/
	return nil
}

// Release is a no op for the default lock service.
func (t *DefaultLockService) Release(obj v1.UserBuildEvent) error {
	return nil
}
