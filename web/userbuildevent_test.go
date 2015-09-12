package main

import "testing"

func TestUserEvent(t *testing.T) {
	event := UserBuildEvent{team: "team", library: "lib", branches: []string{"master"}}
	if event.Team() != "team" {
		t.Fatalf("Want team but got %s\n", event.Team())
	}
	if event.Library() != "lib" {
		t.Fatalf("Want lib but got %s\n", event.Library())
	}
	if event.ProjectKey() != "team/lib" {
		t.Fatalf("Want team/lib but got %s\n", event.ProjectKey())
	}

	branches := event.Branches()
	if len(branches) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(branches))
	}
	if branches[0] != "master" {
		t.Fatalf("Want master but got %s\n", branches[0])
	}
}
