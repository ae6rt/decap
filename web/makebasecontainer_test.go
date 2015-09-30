package main

import "testing"

func TestMakeBaseContainer(t *testing.T) {
	k8s := NewDefaultDecap("url", "admin", "admin123", "key", "sekrit", "us-west-1", NoOpLocker{}, "repo", "repobranch")

	buildEvent := UserBuildEvent{TeamFld: "ae6rt", LibraryFld: "somelib", RefsFld: []string{"master"}}
	baseContainer := k8s.makeBaseContainer(buildEvent, "uuid", "master", map[string]Project{
		"ae6rt/somelib": Project{Team: "ae6rt", Library: "somelib", Descriptor: ProjectDescriptor{Image: "magic-image"}, Sidecars: []string{}},
	})

	if baseContainer.Name != "build-server" {
		t.Fatalf("Want build-server but got %s\n", baseContainer.Name)
	}

	if baseContainer.Image != "magic-image" {
		t.Fatalf("Want magic-image", baseContainer.Image)
	}
}
