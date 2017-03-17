package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/locks"
)

func TestDeferBuild(t *testing.T) {
	locker := locks.NoOpLocker{}
	builder := DefaultBuilder{Locker: &locker}

	if err := builder.DeferBuild(v1.UserBuildEvent{Team_: "t1", Project_: "p1", Refs_: []string{"ignored"}}, "issue/1"); err != nil {
		t.Errorf("Unexpected error: %v\n", err)
	}

	data, _ := json.Marshal(&v1.UserBuildEvent{Team_: "t1", Project_: "p1", Refs_: []string{"issue/1"}})
	if !bytes.Equal(locker.Data, data) {
		t.Fatal("Expecting true")
	}
}
