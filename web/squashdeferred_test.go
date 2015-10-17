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

	expected := map[int]string{
		0: "/1",
		1: "/3",
		2: "/4",
		3: "/6",
		4: "/7",
	}
	for k, v := range expected {
		if squashed[k].Deferral.Key != v {
			t.Fatalf("Want %s but got %s\n", v, squashed[k].Deferral.Key)
		}
	}
}
