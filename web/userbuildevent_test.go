package main

import (
	"encoding/hex"
	"testing"
)

func TestUserEvent(t *testing.T) {
	event := UserBuildEvent{Team_: "team", Project_: "lib", Refs_: []string{"master"}}
	if event.Team() != "team" {
		t.Fatalf("Want team but got %s\n", event.Team())
	}
	if event.Project() != "lib" {
		t.Fatalf("Want lib but got %s\n", event.Project())
	}
	if event.Key() != "team/lib" {
		t.Fatalf("Want team/lib but got %s\n", event.Key())
	}

	if event.DeferralID() != hex.EncodeToString([]byte("team/lib/master")) {
		t.Fatalf("Want %s but got %s\n", hex.EncodeToString([]byte("team/lib/master")), event.DeferralID())
	}

	branches := event.Refs()
	if len(branches) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(branches))
	}
	if branches[0] != "master" {
		t.Fatalf("Want master but got %s\n", branches[0])
	}
}
