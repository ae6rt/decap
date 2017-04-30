package cluster

import k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"

// KubernetesClient is the subset we need of the full client API
type KubernetesClient interface {
	k8sv1.PodsGetter
	k8sv1.SecretsGetter
}
