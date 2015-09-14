package main

type MockDecap struct {
	event BuildEvent
}

func (d *MockDecap) LaunchBuild(p BuildEvent) error {
	d.event = p
	return nil
}

func (d *MockDecap) DeletePod(podName string) error {
	return nil
}
