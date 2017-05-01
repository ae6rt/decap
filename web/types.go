package main

import k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"

// RepoManagerCredential models the username and password for supported source code repository managers, such as Github or Atlassian Stash.
// For Github, the User is the OAuth2 access key and Password is the application's OAuth2 token.
type RepoManagerCredential struct {
	User     string
	Password string
}

// ClusterService models the Kubernetes client interface
// todo is this used?
type ClusterService interface {
}

// BuildScriptsRepo models where the build scripts are held
type BuildScripts struct {
	URL    string
	Branch string
}

// KubernetesClient is the subset we need of the full client API
type KubernetesClient interface {
	k8sv1.PodsGetter
	k8sv1.SecretsGetter
}
