package main

type MockBuilder struct {
	event   BuildEvent
	buildID string
}

func (d *MockBuilder) LaunchBuild(p BuildEvent) error {
	d.event = p
	return nil
}

func (d *MockBuilder) DeletePod(podName string) error {
	d.buildID = podName
	return nil
}

func (d *MockBuilder) DeferBuild(event BuildEvent, branch string) error {
	return nil
}
