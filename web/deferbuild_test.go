package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ae6rt/decap/web/locks"
)

func TestDeferBuild(t *testing.T) {
	locker := locks.NoOpLocker{}
	builder := DefaultBuilder{Locker: &locker}

	builder.DeferBuild(UserBuildEvent{Team_: "t1", Project_: "p1", Refs_: []string{"ignored"}}, "issue/1")

	data, _ := json.Marshal(&UserBuildEvent{Team_: "t1", Project_: "p1", Refs_: []string{"issue/1"}})
	if !bytes.Equal(locker.Data, data) {
		t.Fatal("Expecting true")
	}
}
