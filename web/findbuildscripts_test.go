package main

import (
	"os"
	"testing"

	"github.com/ae6rt/ziptools"
)

func TestFindBuildScripts(t *testing.T) {
	dir, err := ziptools.Unzip("buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}

	scripts, err := findBuildScripts(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	if len(scripts) != 3 {
		os.RemoveAll(dir)
		t.Fatalf("Want 3 but got %d\n", len(scripts))
	}
	os.RemoveAll(dir)
}
