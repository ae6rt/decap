package main

import (
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewKubernetesClient returns a new client.
func NewKubernetesClient() (KubernetesClient, error) {
	var cfg *rest.Config
	var err error

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	return clientset.CoreV1(), err
}

func kubeSecret(file string, defaultValue string) string {
	v, err := ioutil.ReadFile(file)
	if err != nil {
		parts := strings.Split(file, "/")
		Log.Printf("Using command line default %s\n", parts[len(parts)-1])
		return defaultValue
	}
	Log.Printf("Successfully read secret %s from the filesystem\n", file)
	return string(v)
}
