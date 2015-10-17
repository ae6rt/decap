package main

import (
	"testing"

	"github.com/ae6rt/decap/web/locks"
)

func TestSquashDeferred(t *testing.T) {
	deferrals := []locks.Deferral{
		locks.Deferral{
			Data: `{"team": "t1", "project": "p1", "refs": ["master"]}`, // KEEP
			Key:  "/1",
		},
		locks.Deferral{
			Data: `{"team": "t1", "project": "p1", "refs": ["master"]}`, // dup of 0
			Key:  "/2",
		},
		locks.Deferral{
			Data: `{"team": "t2", "project": "x1", "refs": ["develop"]}`, // KEEP
			Key:  "/3",
		},
		locks.Deferral{
			Data: `{"team": "t3", "project": "b1", "refs": ["issue/1"]}`, // KEEP
			Key:  "/4",
		},
		locks.Deferral{
			Data: `{"team": "t3", "project": "b1", "refs": ["issue/1"]}`, // dup of 3
			Key:  "/5",
		},
		locks.Deferral{
			Data: `{"team": "t4", "project": "s1", "refs": ["hotfix/1"]}`, // KEEP
			Key:  "/6",
		},
		locks.Deferral{
			Data: `{"team": "t1", "project": "p1", "refs": ["hotfix/1"]}`, // KEEP
			Key:  "/7",
		},
	}

	squashed := DefaultBuilder{}.SquashDeferred(deferrals)

	if len(squashed) != 5 {
		t.Fatalf("Want 5 but got %d\n", len(squashed))
	}

	i := 0
	if squashed[i].Deferral.Key != "/1" {
		t.Fatalf("Want /1 but got %d\n", squashed[i].Deferral.Key)
	}

	i += 1
	if squashed[i].Deferral.Key != "/3" {
		t.Fatalf("Want /3 but got %d\n", squashed[i].Deferral.Key)
	}

	i += 1
	if squashed[i].Deferral.Key != "/4" {
		t.Fatalf("Want /4 but got %d\n", squashed[i].Deferral.Key)
	}

	i += 1
	if squashed[i].Deferral.Key != "/6" {
		t.Fatalf("Want /6 but got %d\n", squashed[i].Deferral.Key)
	}

	i += 1
	if squashed[i].Deferral.Key != "/7" {
		t.Fatalf("Want /7 but got %d\n", squashed[i].Deferral.Key)
	}
}
