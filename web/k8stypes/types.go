/*
Copyright 2015 The Kubernetes Authors All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8stypes

import "time"

type Pod struct {
	TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the pod.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
	Spec PodSpec `json:"spec,omitempty"`

	// Most recently observed status of the pod.
	// This data may not be up to date.
	// Populated by the system.
	// Read-only.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
	Status PodStatus `json:"status,omitempty"`
}

type RestartPolicy string

type DNSPolicy string

type PodSpec struct {
	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/volumes.md
	Volumes []Volume `json:"volumes,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// List of containers belonging to the pod.
	// Containers cannot currently be added or removed.
	// There must be at least one container in a Pod.
	// Cannot be updated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/containers.md
	Containers []Container `json:"containers" patchStrategy:"merge" patchMergeKey:"name"`
	// Restart policy for all containers within the pod.
	// One of Always, OnFailure, Never.
	// Default to Always.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#restartpolicy
	RestartPolicy RestartPolicy `json:"restartPolicy,omitempty"`
	// Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
	// Value must be non-negative integer. The value zero indicates delete immediately.
	// If this value is nil, the default grace period will be used instead.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	// Set this value longer than the expected cleanup time for your process.
	// Defaults to 30 seconds.
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
	// Optional duration in seconds the pod may be active on the node relative to
	// StartTime before the system will actively try to mark it failed and kill associated containers.
	// Value must be a positive integer.
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`
	// Set DNS policy for containers within the pod.
	// One of 'ClusterFirst' or 'Default'.
	// Defaults to "ClusterFirst".
	DNSPolicy DNSPolicy `json:"dnsPolicy,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/node-selection/README.md
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: http://releases.k8s.io/HEAD/docs/design/service_accounts.md
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// DeprecatedServiceAccount is a depreciated alias for ServiceAccountName.
	// Deprecated: Use serviceAccountName instead.
	DeprecatedServiceAccount string `json:"serviceAccount,omitempty"`

	// NodeName is a request to schedule this pod onto a specific node. If it is non-empty,
	// the scheduler simply schedules this pod onto that node, assuming that it fits resource
	// requirements.
	NodeName string `json:"nodeName,omitempty"`
	// Host networking requested for this pod. Use the host's network namespace.
	// If this option is set, the ports that will be used must be specified.
	// Default to false.
	HostNetwork bool `json:"hostNetwork,omitempty"`
	// Use the host's pid namespace.
	// Optional: Default to false.
	HostPID bool `json:"hostPID,omitempty"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use. For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/images.md#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

type PodPhase string

const (
	// PodPending means the pod has been accepted by the system, but one or more of the containers
	// has not been started. This includes time before being bound to a node, as well as time spent
	// pulling images onto the host.
	PodPending PodPhase = "Pending"
	// PodRunning means the pod has been bound to a node and all of the containers have been started.
	// At least one container is still running or is in the process of being restarted.
	PodRunning PodPhase = "Running"
	// PodSucceeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	PodSucceeded PodPhase = "Succeeded"
	// PodFailed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	PodFailed PodPhase = "Failed"
	// PodUnknown means that for some reason the state of the pod could not be obtained, typically due
	// to an error in communicating with the host of the pod.
	PodUnknown PodPhase = "Unknown"
)

type PodCondition struct {
	// Type is the type of the condition.
	// Currently only Ready.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#pod-conditions
	Type PodConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#pod-conditions
	Status ConditionStatus `json:"status"`
	// Last time we probed the condition.
	LastProbeTime Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

type Time struct {
	time.Time
}

type PodConditionType string

const (
	// PodReady means the pod is able to service requests and should be added to the
	// load balancing pools of all matching services.
	PodReady PodConditionType = "Ready"
)

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type ContainerState struct {
	// Details about a waiting container
	Waiting *ContainerStateWaiting `json:"waiting,omitempty"`
	// Details about a running container
	Running *ContainerStateRunning `json:"running,omitempty"`
	// Details about a terminated container
	Terminated *ContainerStateTerminated `json:"terminated,omitempty"`
}

type ContainerStateWaiting struct {
	// (brief) reason the container is not yet running.
	Reason string `json:"reason,omitempty"`
	// Message regarding why the container is not yet running.
	Message string `json:"message,omitempty"`
}

type ContainerStateRunning struct {
	// Time at which the container was last (re-)started
	StartedAt Time `json:"startedAt,omitempty"`
}

type ContainerStateTerminated struct {
	// Exit status from the last termination of the container
	ExitCode int `json:"exitCode"`
	// Signal from the last termination of the container
	Signal int `json:"signal,omitempty"`
	// (brief) reason from the last termination of the container
	Reason string `json:"reason,omitempty"`
	// Message regarding the last termination of the container
	Message string `json:"message,omitempty"`
	// Time at which previous execution of the container started
	StartedAt Time `json:"startedAt,omitempty"`
	// Time at which the container last terminated
	FinishedAt Time `json:"finishedAt,omitempty"`
	// Container's ID in the format 'docker://<container_id>'
	ContainerID string `json:"containerID,omitempty"`
}

type ContainerStatus struct {
	// This must be a DNS_LABEL. Each container in a pod must have a unique name.
	// Cannot be updated.
	Name string `json:"name"`
	// Details about the container's current condition.
	State ContainerState `json:"state,omitempty"`
	// Details about the container's last termination condition.
	LastTerminationState ContainerState `json:"lastState,omitempty"`
	// Specifies whether the container has passed its readiness probe.
	Ready bool `json:"ready"`
	// The number of times the container has been restarted, currently based on
	// the number of dead containers that have not yet been removed.
	// Note that this is calculated from dead containers. But those containers are subject to
	// garbage collection. This value will get capped at 5 by GC.
	RestartCount int `json:"restartCount"`
	// The image the container is running.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/images.md
	// TODO(dchen1107): Which image the container is running with?
	Image string `json:"image"`
	// ImageID of the container's image.
	ImageID string `json:"imageID"`
	// Container's ID in the format 'docker://<container_id>'.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/container-environment.md#container-information
	ContainerID string `json:"containerID,omitempty"`
}

type PodStatus struct {
	// Current condition of the pod.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#pod-phase
	Phase PodPhase `json:"phase,omitempty"`
	// Current service state of pod.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#pod-conditions
	Conditions []PodCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// A human readable message indicating details about why the pod is in this condition.
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the pod is in this state.
	// e.g. 'OutOfDisk'
	Reason string `json:"reason,omitempty"`

	// IP address of the host to which the pod is assigned. Empty if not yet scheduled.
	HostIP string `json:"hostIP,omitempty"`
	// IP address allocated to the pod. Routable at least within the cluster.
	// Empty if not yet allocated.
	PodIP string `json:"podIP,omitempty"`

	// RFC 3339 date and time at which the object was acknowledged by the Kubelet.
	// This is before the Kubelet pulled the container image(s) for the pod.
	StartTime *Time `json:"startTime,omitempty"`

	// The list has one entry per container in the manifest. Each entry is currently the output
	// of `docker inspect`.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#container-statuses
	ContainerStatuses []ContainerStatus `json:"containerStatuses,omitempty"`
}

type TypeMeta struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#types-kinds
	Kind string `json:"kind,omitempty"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#resources
	APIVersion string `json:"apiVersion,omitempty"`
}

type ObjectMeta struct {
	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/identifiers.md#names
	Name string `json:"name,omitempty"`

	// GenerateName is an optional prefix, used by the server, to generate a unique
	// name ONLY IF the Name field has not been provided.
	// If this field is used, the name returned to the client will be different
	// than the name passed. This value will also be combined with a unique suffix.
	// The provided value has the same validation rules as the Name field,
	// and may be truncated by the length of the suffix required to make the value
	// unique on the server.
	//
	// If this field is specified and the generated name exists, the server will
	// NOT return a 409 - instead, it will either return 201 Created or 500 with Reason
	// ServerTimeout indicating a unique name could not be found in the time allotted, and the client
	// should retry (optionally after the time indicated in the Retry-After header).
	//
	// Applied only if Name is not specified.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#idempotency
	GenerateName string `json:"generateName,omitempty"`

	// Namespace defines the space within each name must be unique. An empty namespace is
	// equivalent to the "default" namespace, but "default" is the canonical representation.
	// Not all objects are required to be scoped to a namespace - the value of this field for
	// those objects will be empty.
	//
	// Must be a DNS_LABEL.
	// Cannot be updated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/namespaces.md
	Namespace string `json:"namespace,omitempty"`

	// SelfLink is a URL representing this object.
	// Populated by the system.
	// Read-only.
	SelfLink string `json:"selfLink,omitempty"`

	// UID is the unique in time and space value for this object. It is typically generated by
	// the server on successful creation of a resource and is not allowed to change on PUT
	// operations.
	//
	// Populated by the system.
	// Read-only.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/identifiers.md#uids
	UID string `json:"uid,omitempty"`

	// An opaque value that represents the internal version of this object that can
	// be used by clients to determine when objects have changed. May be used for optimistic
	// concurrency, change detection, and the watch operation on a resource or set of resources.
	// Clients must treat these values as opaque and passed unmodified back to the server.
	// They may only be valid for a particular resource or set of resources.
	//
	// Populated by the system.
	// Read-only.
	// Value must be treated as opaque by clients and .
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#concurrency-control-and-consistency
	ResourceVersion string `json:"resourceVersion,omitempty"`

	// A sequence number representing a specific generation of the desired state.
	// Currently only implemented by replication controllers.
	// Populated by the system.
	// Read-only.
	Generation int64 `json:"generation,omitempty"`

	// CreationTimestamp is a timestamp representing the server time when this object was
	// created. It is not guaranteed to be set in happens-before order across separate operations.
	// Clients may not set this value. It is represented in RFC3339 form and is in UTC.
	//
	// Populated by the system.
	// Read-only.
	// Null for lists.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	CreationTimestamp Time `json:"creationTimestamp,omitempty"`

	// DeletionTimestamp is RFC 3339 date and time at which this resource will be deleted. This
	// field is set by the server when a graceful deletion is requested by the user, and is not
	// directly settable by a client. The resource will be deleted (no longer visible from
	// resource lists, and not reachable by name) after the time in this field. Once set, this
	// value may not be unset or be set further into the future, although it may be shortened
	// or the resource may be deleted prior to this time. For example, a user may request that
	// a pod is deleted in 30 seconds. The Kubelet will react by sending a graceful termination
	// signal to the containers in the pod. Once the resource is deleted in the API, the Kubelet
	// will send a hard termination signal to the container.
	// If not set, graceful deletion of the object has not been requested.
	//
	// Populated by the system when a graceful deletion is requested.
	// Read-only.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	DeletionTimestamp *Time `json:"deletionTimestamp,omitempty"`

	// Number of seconds allowed for this object to gracefully terminate before
	// it will be removed from the system. Only set when deletionTimestamp is also set.
	// May only be shortened.
	// Read-only.
	DeletionGracePeriodSeconds *int64 `json:"deletionGracePeriodSeconds,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/labels.md
	// TODO: replace map[string]string with labels.LabelSet type
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/annotations.md
	Annotations map[string]string `json:"annotations,omitempty"`
}

type LocalObjectReference struct {
	// Name of the referent.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/identifiers.md#names
	// TODO: Add other useful fields. apiVersion, kind, uid?
	Name string `json:"name,omitempty"`
}

type Volume struct {
	// Volume's name.
	// Must be a DNS_LABEL and unique within the pod.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/identifiers.md#names
	Name string `json:"name"`
	// VolumeSource represents the location and type of the mounted volume.
	// If not specified, the Volume is implied to be an EmptyDir.
	// This implied behavior is deprecated and will be removed in a future version.
	VolumeSource `json:",inline"`
}

type VolumeSource struct {
	// HostPath represents a pre-existing file or directory on the host
	// machine that is directly exposed to the container. This is generally
	// used for system agents or other privileged things that are allowed
	// to see the host machine. Most containers will NOT need this.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/volumes.md#hostpath
	// ---
	// TODO(jonesdl) We need to restrict who can use host directory mounts and who can/can not
	// mount host directories as read/write.
	HostPath *HostPathVolumeSource `json:"hostPath,omitempty"`
	// EmptyDir represents a temporary directory that shares a pod's lifetime.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/volumes.md#emptydir
	EmptyDir *EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// GitRepo represents a git repository at a particular revision.
	GitRepo *GitRepoVolumeSource `json:"gitRepo,omitempty"`
	// Secret represents a secret that should populate this volume.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/volumes.md#secrets
	Secret *SecretVolumeSource `json:"secret,omitempty"`
}

type SecretVolumeSource struct {
	// SecretName is the name of a secret in the pod's namespace.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/volumes.md#secrets
	SecretName string `json:"secretName"`
}

type HostPathVolumeSource struct {
	// Path of the directory on the host.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/volumes.md#hostpath
	Path string `json:"path"`
}

type EmptyDirVolumeSource struct {
	// What type of storage medium should back this directory.
	// The default is "" which means to use the node's default medium.
	// Must be an empty string (default) or Memory.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/volumes.md#emptydir
	Medium StorageMedium `json:"medium,omitempty"`
}

type StorageMedium string

const (
	StorageMediumDefault StorageMedium = ""       // use whatever the default is for the node
	StorageMediumMemory  StorageMedium = "Memory" // use memory (tmpfs)
)

type GitRepoVolumeSource struct {
	// Repository URL
	Repository string `json:"repository"`
	// Commit hash for the specified revision.
	Revision string `json:"revision"`
}

type Container struct {
	// Name of the container specified as a DNS_LABEL.
	// Each container in a pod must have a unique name (DNS_LABEL).
	// Cannot be updated.
	Name string `json:"name"`
	// Docker image name.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/images.md
	Image string `json:"image,omitempty"`
	// Entrypoint array. Not executed within a shell.
	// The docker image's entrypoint is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax
	// can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded,
	// regardless of whether the variable exists or not.
	// Cannot be updated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/containers.md#containers-and-commands
	Command []string `json:"command,omitempty"`
	// Arguments to the entrypoint.
	// The docker image's cmd is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax
	// can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded,
	// regardless of whether the variable exists or not.
	// Cannot be updated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/containers.md#containers-and-commands
	Args []string `json:"args,omitempty"`
	// Container's working directory.
	// Defaults to Docker's default. D
	// efaults to image's default.
	// Cannot be updated.
	WorkingDir string `json:"workingDir,omitempty"`
	// List of ports to expose from the container.
	// Cannot be updated.
	Ports []ContainerPort `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"containerPort"`
	// List of environment variables to set in the container.
	// Cannot be updated.
	Env []EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/persistent-volumes.md#resources
	//Resources ResourceRequirements `json:"resources,omitempty"`
	// Pod volumes to mount into the container's filesyste.
	// Cannot be updated.
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// Periodic probe of container liveness.
	// Container will be restarted if the probe fails.
	// Cannot be updated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#container-probes
	LivenessProbe *Probe `json:"livenessProbe,omitempty"`
	// Periodic probe of container service readiness.
	// Container will be removed from service endpoints if the probe fails.
	// Cannot be updated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#container-probes
	ReadinessProbe *Probe `json:"readinessProbe,omitempty"`
	// Actions that the management system should take in response to container lifecycle events.
	// Cannot be updated.
	Lifecycle *Lifecycle `json:"lifecycle,omitempty"`
	// Optional: Path at which the file to which the container's termination message
	// will be written is mounted into the container's filesystem.
	// Message written is intended to be brief final status, such as an assertion failure message.
	// Defaults to /dev/termination-log.
	// Cannot be updated.
	TerminationMessagePath string `json:"terminationMessagePath,omitempty"`
	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/images.md#updating-images
	ImagePullPolicy PullPolicy `json:"imagePullPolicy,omitempty"`
	// Security options the pod should run with.
	// More info: http://releases.k8s.io/HEAD/docs/design/security_context.md
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`

	// Whether this container should allocate a buffer for stdin in the container runtime.
	// Default is false.
	Stdin bool `json:"stdin,omitempty"`
	// Whether this container should allocate a TTY for itself, also requires 'stdin' to be true.
	// Default is false.
	TTY bool `json:"tty,omitempty"`
}

type ContainerPort struct {
	// If specified, this must be an IANA_SVC_NAME and unique within the pod. Each
	// named port in a pod must have a unique name. Name for the port that can be
	// referred to by services.
	Name string `json:"name,omitempty"`
	// Number of port to expose on the host.
	// If specified, this must be a valid port number, 0 < x < 65536.
	// If HostNetwork is specified, this must match ContainerPort.
	// Most containers do not need this.
	HostPort int `json:"hostPort,omitempty"`
	// Number of port to expose on the pod's IP address.
	// This must be a valid port number, 0 < x < 65536.
	ContainerPort int `json:"containerPort"`
	// Protocol for port. Must be UDP or TCP.
	// Defaults to "TCP".
	Protocol Protocol `json:"protocol,omitempty"`
	// What host IP to bind the external port to.
	HostIP string `json:"hostIP,omitempty"`
}

type Protocol string

const (
	// ProtocolTCP is the TCP protocol.
	ProtocolTCP Protocol = "TCP"
	// ProtocolUDP is the UDP protocol.
	ProtocolUDP Protocol = "UDP"
)

type EnvVar struct {
	// Name of the environment variable. Must be a C_IDENTIFIER.
	Name string `json:"name"`

	// Variable references $(VAR_NAME) are expanded
	// using the previous defined environment variables in the container and
	// any service environment variables. If a variable cannot be resolved,
	// the reference in the input string will be unchanged. The $(VAR_NAME)
	// syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped
	// references will never be expanded, regardless of whether the variable
	// exists or not.
	// Defaults to "".
	Value string `json:"value,omitempty"`
	// Source for the environment variable's value. Cannot be used if value is not empty.
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

type EnvVarSource struct {
	// Selects a field of the pod. Only name and namespace are supported.
	FieldRef *ObjectFieldSelector `json:"fieldRef"`
}

type ObjectFieldSelector struct {
	// Version of the schema the FieldPath is written in terms of, defaults to "v1".
	APIVersion string `json:"apiVersion,omitempty"`
	// Path of the field to select in the specified API version.
	FieldPath string `json:"fieldPath"`
}

type VolumeMount struct {
	// This must match the Name of a Volume.
	Name string `json:"name"`
	// Mounted read-only if true, read-write otherwise (false or unspecified).
	// Defaults to false.
	ReadOnly bool `json:"readOnly,omitempty"`
	// Path within the container at which the volume should be mounted.
	MountPath string `json:"mountPath"`
}

type Probe struct {
	// The action taken to determine the health of a container
	Handler `json:",inline"`
	// Number of seconds after the container has started before liveness probes are initiated.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#container-probes
	InitialDelaySeconds int64 `json:"initialDelaySeconds,omitempty"`
	// Number of seconds after which liveness probes timeout.
	// Defaults to 1 second.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/pod-states.md#container-probes
	TimeoutSeconds int64 `json:"timeoutSeconds,omitempty"`
}

type Handler struct {
	// One and only one of the following should be specified.
	// Exec specifies the action to take.
	Exec *ExecAction `json:"exec,omitempty"`
	// HTTPGet specifies the http request to perform.
	HTTPGet *HTTPGetAction `json:"httpGet,omitempty"`
	// TCPSocket specifies an action involving a TCP port.
	// TCP hooks not yet supported
	// TODO: implement a realistic TCP lifecycle hook
	TCPSocket *TCPSocketAction `json:"tcpSocket,omitempty"`
}

type ExecAction struct {
	// Command is the command line to execute inside the container, the working directory for the
	// command  is root ('/') in the container's filesystem. The command is simply exec'd, it is
	// not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use
	// a shell, you need to explicitly call out to that shell.
	// Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
	Command []string `json:"command,omitempty"`
}

type HTTPGetAction struct {
	// Path to access on the HTTP server.
	Path string `json:"path,omitempty"`
	// Name or number of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port int `json:"port"`
	// Host name to connect to, defaults to the pod IP.
	Host string `json:"host,omitempty"`
	// Scheme to use for connecting to the host.
	// Defaults to HTTP.
	Scheme URIScheme `json:"scheme,omitempty"`
}

type URIScheme string

const (
	// URISchemeHTTP means that the scheme used will be http://
	URISchemeHTTP URIScheme = "HTTP"
	// URISchemeHTTPS means that the scheme used will be https://
	URISchemeHTTPS URIScheme = "HTTPS"
)

type TCPSocketAction struct {
	// Number or name of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port int `json:"port"`
}

type Lifecycle struct {
	// PostStart is called immediately after a container is created. If the handler fails,
	// the container is terminated and restarted according to its restart policy.
	// Other management of the container blocks until the hook completes.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/container-environment.md#hook-details
	PostStart *Handler `json:"postStart,omitempty"`
	// PreStop is called immediately before a container is terminated.
	// The container is terminated after the handler completes.
	// The reason for termination is passed to the handler.
	// Regardless of the outcome of the handler, the container is eventually terminated.
	// Other management of the container blocks until the hook completes.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/container-environment.md#hook-details
	PreStop *Handler `json:"preStop,omitempty"`
}

type PullPolicy string

const (
	// PullAlways means that kubelet always attempts to pull the latest image. Container will fail If the pull fails.
	PullAlways PullPolicy = "Always"
	// PullNever means that kubelet never pulls an image, but only uses a local image. Container will fail if the image isn't present
	PullNever PullPolicy = "Never"
	// PullIfNotPresent means that kubelet pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.
	PullIfNotPresent PullPolicy = "IfNotPresent"
)

type SecurityContext struct {
	// The linux kernel capabilites that should be added or removed.
	// Default to Container.Capabilities if left unset.
	// More info: http://releases.k8s.io/HEAD/docs/design/security_context.md#security-context
	Capabilities *Capabilities `json:"capabilities,omitempty"`

	// Run the container in privileged mode.
	// Default to Container.Privileged if left unset.
	// More info: http://releases.k8s.io/HEAD/docs/design/security_context.md#security-context
	Privileged *bool `json:"privileged,omitempty"`

	// SELinuxOptions are the labels to be applied to the container
	// and volumes.
	// Options that control the SELinux labels applied.
	// More info: http://releases.k8s.io/HEAD/docs/design/security_context.md#security-context
	SELinuxOptions *SELinuxOptions `json:"seLinuxOptions,omitempty"`

	// RunAsUser is the UID to run the entrypoint of the container process.
	// The user id that runs the first process in the container.
	// More info: http://releases.k8s.io/HEAD/docs/design/security_context.md#security-context
	RunAsUser *int64 `json:"runAsUser,omitempty"`

	// RunAsNonRoot indicates that the container should be run as a non-root user. If the RunAsUser
	// field is not explicitly set then the kubelet may check the image for a specified user or
	// perform defaulting to specify a user.
	RunAsNonRoot bool `json:"runAsNonRoot,omitempty"`
}

type Capabilities struct {
	// Added capabilities
	Add []Capability `json:"add,omitempty"`
	// Removed capabilities
	Drop []Capability `json:"drop,omitempty"`
}

type Capability string

type SELinuxOptions struct {
	// User is a SELinux user label that applies to the container.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/labels.md
	User string `json:"user,omitempty"`

	// Role is a SELinux role label that applies to the container.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/labels.md
	Role string `json:"role,omitempty"`

	// Type is a SELinux type label that applies to the container.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/labels.md
	Type string `json:"type,omitempty"`

	// Level is SELinux level label that applies to the container.
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/labels.md
	Level string `json:"level,omitempty"`
}
