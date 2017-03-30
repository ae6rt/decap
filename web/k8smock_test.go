package main

import (
	"fmt"
	"testing"

	k8s2 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/policy/v1beta1"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/rest"
)

type podLifter struct {
}

func (t podLifter) Create(pod *v1.Pod) (*v1.Pod, error) {
	return nil, nil
}

func (t podLifter) Update(pod *v1.Pod) (*v1.Pod, error) {
	return nil, nil
}

func (t podLifter) UpdateStatus(pod *v1.Pod) (*v1.Pod, error) {
	return nil, nil
}

func (t podLifter) Delete(name string, options *v1.DeleteOptions) error {
	return nil
}

func (t podLifter) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return nil
}

func (t podLifter) Get(name string) (*v1.Pod, error) {
	return nil, nil
}

func (t podLifter) List(options v1.ListOptions) (*v1.PodList, error) {
	return nil, nil
}

func (t podLifter) Watch(options v1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (t podLifter) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.Pod, err error) {
	return nil, nil
}

func (t podLifter) Bind(binding *v1.Binding) error {
	return nil
}

func (t podLifter) Evict(eviction *v1beta1.Eviction) error {
	return nil
}

func (t podLifter) GetLogs(name string, opts *v1.PodLogOptions) *rest.Request {
	return nil
}

type podGetter struct {
}

func (t podGetter) Pods(namespace string) k8s2.PodInterface {
	return podLifter{}
}

func TestK8sMock(t *testing.T) {
	buildLauncher := NewBuildLauncher("", "", nil, nil, podGetter{}, nil)
	fmt.Println(buildLauncher)
}
