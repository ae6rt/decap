package v1

import "testing"

func TestUserEvent(t *testing.T) {
	event := UserBuildEvent{Team_: "team", Project_: "lib", Ref_: "master"}
	if event.Team_ != "team" {
		t.Fatalf("Want team but got %s\n", event.Team_)
	}
	if event.Project_ != "lib" {
		t.Fatalf("Want lib but got %s\n", event.Project_)
	}
	if event.Lockname() != "team/lib/master" {
		t.Fatalf("Want team/lib/master but got %s\n", event.Lockname())
	}

	branch := event.Ref_
	if branch != "master" {
		t.Fatalf("Want master but got %s\n", branch)
	}
}
