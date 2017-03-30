package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// NewKubernetesClient returns a new client for use inside the cluster.
// TODO extend for use outside the cluster
func NewKubernetesClient() (k8sv1.PodsGetter, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset.CoreV1(), err
}
