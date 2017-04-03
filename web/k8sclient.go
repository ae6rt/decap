package main

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewKubernetesClient returns a new client.
func NewKubernetesClient() (KubernetesClient, error) {

	var cfg *rest.Config
	var err error

	cfg, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	if err != nil {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	return clientset.CoreV1(), err
}
