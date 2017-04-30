package main

import (
	"time"

	k8scorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/pkg/api"
	k8sapiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/policy/v1beta1"
	"k8s.io/client-go/pkg/watch"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/deferrals"
	k8sapi "k8s.io/client-go/pkg/api/v1"
)

// BuildManager
type BuildManagerBaseMock struct {
}

func (d *BuildManagerBaseMock) LaunchBuild(p v1.UserBuildEvent) error {
	return nil
}

func (d *BuildManagerBaseMock) DeletePod(podName string) error {
	return nil
}

func (d *BuildManagerBaseMock) DeferBuild(event deferrals.Deferrable) error {
	return nil
}

func (d *BuildManagerBaseMock) DeferredBuilds() ([]deferrals.Deferrable, error) {
	return nil, nil
}

func (d *BuildManagerBaseMock) CreatePod(pod *k8sapi.Pod) error {
	return nil
}

func (d *BuildManagerBaseMock) PodWatcher() {
}

func (d *BuildManagerBaseMock) LaunchDeferred(ticker <-chan time.Time) {
}

func (d *BuildManagerBaseMock) ClearDeferredBuild(key string) error {
	return nil
}

func (d *BuildManagerBaseMock) QueueIsOpen() bool {
	return true
}

func (d *BuildManagerBaseMock) OpenQueue() {
}

func (d *BuildManagerBaseMock) CloseQueue() {
}

// ProjectManager
type ProjectManagerBaseMock struct {
}

func (t *ProjectManagerBaseMock) Assemble() error {
	return nil
}

func (t *ProjectManagerBaseMock) Get(key string) *v1.Project {
	return nil
}

func (t *ProjectManagerBaseMock) GetProjects() map[string]v1.Project {
	return nil
}

func (t *ProjectManagerBaseMock) GetProjectByTeamName(team, projectName string) (v1.Project, bool) {
	return v1.Project{}, true
}

func (t *ProjectManagerBaseMock) RepositoryURL() string {
	return ""
}

func (t *ProjectManagerBaseMock) RepositoryBranch() string {
	return ""
}

// DeferralService

type DeferralServiceBaseMock struct {
}

func (t DeferralServiceBaseMock) Defer(event deferrals.Deferrable) error {
	return nil
}

func (t DeferralServiceBaseMock) Poll() ([]deferrals.Deferrable, error) {
	return nil, nil
}

func (t DeferralServiceBaseMock) List() ([]deferrals.Deferrable, error) {
	return nil, nil
}

func (t DeferralServiceBaseMock) Remove(id string) error {
	return nil
}

// LockService

type LockserviceBaseMock struct {
}

func (t LockserviceBaseMock) Acquire(event v1.UserBuildEvent) error {
	return nil
}

func (t LockserviceBaseMock) Release(event v1.UserBuildEvent) error {
	return nil
}

// Kubernetes client

type podOps struct {
}

func (t podOps) Create(pod *k8sapiv1.Pod) (*k8sapiv1.Pod, error) {
	return nil, nil
}

func (t podOps) Update(pod *k8sapiv1.Pod) (*k8sapiv1.Pod, error) {
	return nil, nil
}

func (t podOps) UpdateStatus(pod *k8sapiv1.Pod) (*k8sapiv1.Pod, error) {
	return nil, nil
}

func (t podOps) Delete(name string, options *k8sapiv1.DeleteOptions) error {
	return nil
}

func (t podOps) DeleteCollection(options *k8sapiv1.DeleteOptions, listOptions k8sapiv1.ListOptions) error {
	return nil
}

func (t podOps) Get(name string) (*k8sapiv1.Pod, error) {
	return nil, nil
}

func (t podOps) List(options k8sapiv1.ListOptions) (*k8sapiv1.PodList, error) {
	return nil, nil
}

func (t podOps) Watch(options k8sapiv1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (t podOps) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *k8sapiv1.Pod, err error) {
	return nil, nil
}

func (t podOps) Bind(binding *k8sapiv1.Binding) error {
	return nil
}

func (t podOps) Evict(eviction *v1beta1.Eviction) error {
	return nil
}

func (t podOps) GetLogs(name string, opts *k8sapiv1.PodLogOptions) *rest.Request {
	return nil
}

type podGetter struct {
}

func (t podGetter) Pods(namespace string) k8scorev1.PodInterface {
	return podOps{}
}

// ##############
// Secrets Getter
// ##############

type secretsGetter struct {
}

func (t secretsGetter) Secrets(namespace string) k8scorev1.SecretInterface {
	return secretOps{}
}

type secretOps struct {
}

func (t secretOps) Create(*k8sapiv1.Secret) (*k8sapiv1.Secret, error) {
	return nil, nil
}

func (t secretOps) Update(*k8sapiv1.Secret) (*k8sapiv1.Secret, error) {
	return nil, nil
}

func (t secretOps) Delete(name string, options *k8sapiv1.DeleteOptions) error {
	return nil
}

func (t secretOps) DeleteCollection(options *k8sapiv1.DeleteOptions, listOptions k8sapiv1.ListOptions) error {
	return nil
}

func (t secretOps) Get(name string) (*k8sapiv1.Secret, error) {
	return nil, nil
}

func (t secretOps) List(options k8sapiv1.ListOptions) (*k8sapiv1.SecretList, error) {
	return nil, nil
}

func (t secretOps) Watch(options k8sapiv1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (t secretOps) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *k8sapiv1.Secret, err error) {
	return nil, nil
}

type mockK8sClient struct {
	podGetter
	secretsGetter
}

type KubernetesClientBaseMock struct {
	podGetter
	secretsGetter
}
