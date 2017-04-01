package main

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// NewKubernetesClient returns a new client.
func NewKubernetesClient() (k8sv1.PodsGetter, error) {

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
