package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewKubernetesClient returns a new client for use inside the cluster.
// TODO extend for use outside the cluster
func NewKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}
