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
	if len(scripts) != 2 {
		os.RemoveAll(dir)
		t.Fatalf("Want 2 but got %d\n", len(scripts))
	}
	os.RemoveAll(dir)
}

func TestFindBuildScriptsByRegex(t *testing.T) {
	dir, err := ziptools.Unzip("buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}

	files, err := filesByRegex(dir, buildScriptRegex)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	if len(files) != 2 {
		os.RemoveAll(dir)
		t.Fatalf("Want 2 but got %d\n", len(files))
	}
	os.RemoveAll(dir)
}

func TestFindProjectDescriptorsByRegex(t *testing.T) {
	dir, err := ziptools.Unzip("buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}

	files, err := filesByRegex(dir, projectDescriptorRegex)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	if len(files) != 1 {
		os.RemoveAll(dir)
		t.Fatalf("Want 1 but got %d\n", len(files))
	}
	os.RemoveAll(dir)
}

func TestFindSidecarsByRegex(t *testing.T) {
	dir, err := ziptools.Unzip("buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}

	files, err := filesByRegex(dir, sideCarRegex)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	if len(files) != 3 {
		os.RemoveAll(dir)
		t.Fatalf("Want 3 but got %d\n", len(files))
	}
	os.RemoveAll(dir)
}

func TestFindByRegexBadRoot(t *testing.T) {
	_, err := filesByRegex("x", "")
	if err == nil {
		t.Fatal("Expecting an error because root is not absolute\n")
	}
}
