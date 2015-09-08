package main

import "testing"

func TestUserEvent(t *testing.T) {
	event := UserBuildEvent{parent: "parent", library: "lib", branches: []string{"master"}}
	if event.Parent() != "parent" {
		t.Fatalf("Want parent but got %s\n", event.Parent())
	}
	if event.Library() != "lib" {
		t.Fatalf("Want lib but got %s\n", event.Library())
	}
	if event.ProjectKey() != "parent/lib" {
		t.Fatalf("Want parent/lib but got %s\n", event.ProjectKey())
	}

	branches := event.Branches()
	if len(branches) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(branches))
	}
	if branches[0] != "master" {
		t.Fatalf("Want master but got %s\n", branches[0])
	}
}
