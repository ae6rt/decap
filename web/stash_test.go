package main

import (
	"testing"
)

func TestStashEvent(t *testing.T) {
	stashEvent := StashEvent{
		Repository: StashRepository{
			Slug: "slug",
			Project: StashProject{
				Key: "proj",
			},
		},
		RefChanges: []StashRefChange{
			StashRefChange{RefID: "refs/heads/master"},
			StashRefChange{RefID: "refs/heads/feature/foo"},
		},
	}
	pushEvent := PushEvent(stashEvent)
	if pushEvent.ProjectKey() != "proj/slug" {
		t.Fatalf("Want proj/slug but got %s\n", pushEvent.ProjectKey())
	}

	branches := pushEvent.Branches()
	if len(branches) != 2 {
		t.Fatalf("Want 2 but got %d\n", len(branches))
	}
	for _, branch := range branches {
		if !(branch == "master" || branch == "feature/foo") {
			t.Fatalf("Want master or feature/foo but got %s\n", branch)
		}
	}
}
