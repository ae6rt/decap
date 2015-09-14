package main

type MockDecap struct {
	event   BuildEvent
	buildID string
}

func (d *MockDecap) LaunchBuild(p BuildEvent) error {
	d.event = p
	return nil
}

func (d *MockDecap) DeletePod(podName string) error {
	d.buildID = podName
	return nil
}
