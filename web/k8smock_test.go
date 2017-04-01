package main

import (
	k8s2 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/policy/v1beta1"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/rest"
)

// Mock the Kubernetes PodGetter interface

type podOps struct {
}

func (t podOps) Create(pod *v1.Pod) (*v1.Pod, error) {
	return nil, nil
}

func (t podOps) Update(pod *v1.Pod) (*v1.Pod, error) {
	return nil, nil
}

func (t podOps) UpdateStatus(pod *v1.Pod) (*v1.Pod, error) {
	return nil, nil
}

func (t podOps) Delete(name string, options *v1.DeleteOptions) error {
	return nil
}

func (t podOps) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return nil
}

func (t podOps) Get(name string) (*v1.Pod, error) {
	return nil, nil
}

func (t podOps) List(options v1.ListOptions) (*v1.PodList, error) {
	return nil, nil
}

func (t podOps) Watch(options v1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (t podOps) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.Pod, err error) {
	return nil, nil
}

func (t podOps) Bind(binding *v1.Binding) error {
	return nil
}

func (t podOps) Evict(eviction *v1beta1.Eviction) error {
	return nil
}

func (t podOps) GetLogs(name string, opts *v1.PodLogOptions) *rest.Request {
	return nil
}

type podGetter struct {
}

func (t podGetter) Pods(namespace string) k8s2.PodInterface {
	return podOps{}
}
