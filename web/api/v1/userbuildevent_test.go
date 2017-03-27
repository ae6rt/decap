package v1

import "testing"

func TestUserEvent(t *testing.T) {
	event := UserBuildEvent{Team: "team", Project: "lib", Ref: "master"}
	if event.Team != "team" {
		t.Fatalf("Want team but got %s\n", event.Team)
	}
	if event.Project != "lib" {
		t.Fatalf("Want lib but got %s\n", event.Project)
	}
	if event.Lockname() != "team/lib/master" {
		t.Fatalf("Want team/lib/master but got %s\n", event.Lockname())
	}

	branch := event.Ref
	if branch != "master" {
		t.Fatalf("Want master but got %s\n", branch)
	}
}
