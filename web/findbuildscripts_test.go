package main

import (
	"fmt"
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

func TestReadSidecars(t *testing.T) {
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

	arr := readSidecars(files)
	if len(arr) != 3 {
		os.RemoveAll(dir)
		t.Fatalf("Want 3 but got %d\n", len(files))
	}

	fmt.Println(arr)
	os.RemoveAll(dir)
}

func TestFindByRegexBadRoot(t *testing.T) {
	_, err := filesByRegex("x", "")
	if err == nil {
		t.Fatal("Expecting an error because root is not absolute\n")
	}
}

func TestProjectKey(t *testing.T) {
	r := projectKey("a", "b")
	if r != "a/b" {
		t.Fatalf("Want a/b but got %s\n", r)
	}
}

func TestTeamProjectFromFile(t *testing.T) {
	a, b, err := teamProject("/a/b/c/d/x.sh")
	if err != nil {
		t.Fatalf("Unexpected error:  %v\n", err)
	}
	if !(a == "c" && b == "d") {
		t.Fatalf("Want c and d but got %s and %s\n", a, b)
	}

	a, b, err = teamProject("/x.sh")
	if err == nil {
		t.Fatalf("Expected an error becuase path is not at least 3 deep\n")
	}
}

func TestIndexFilesByTeamProject(t *testing.T) {
	flist := []string{
		"/a/b/c/d/build.sh",
		"/1/2/3/4/build.sh",
		"/5/6/7/8/build.sh",
	}

	m := indexFilesByTeamProject(flist)
	if len(m) != 3 {
		t.Fatalf("Want 3 but got %d\n", len(m))
	}

	key := "c/d"
	if m[key] != "/a/b/c/d/build.sh" {
		t.Fatalf("Want /a/b/c/d/build.sh but got %s and %s\n", m[key])
	}

	key = "3/4"
	if m[key] != "/1/2/3/4/build.sh" {
		t.Fatalf("Want /1/2/3/4/build.sh but got %s and %s\n", m[key])
	}

	key = "7/8"
	if m[key] != "/5/6/7/8/build.sh" {
		t.Fatalf("Want /5/6/7/8/build.sh but got %s and %s\n", m[key])
	}
}

func TestIndexSidecarsByTeamProject(t *testing.T) {
	flist := []string{
		"/a/b/c/d/mysql-sidecar.json",
		"/a/b/c/d/redis-sidecar.json",
		"/e/f/g/h/redis-sidecar.json",
	}

	m := indexSidecarsByTeamProject(flist)
	if len(m) != 2 {
		t.Fatalf("Want 2 but got %d\n", len(m))
	}

	key := "c/d"
	if len(m[key]) != 2 {
		t.Fatalf("Want 2 but got %d\n", len(m[key]))
	}
	arr := m[key]
	if len(arr) != 2 {
		t.Fatalf("Want 2 but got %d\n", len(arr))
	}
	for _, v := range arr {
		if !(v == "/a/b/c/d/mysql-sidecar.json" || v == "/a/b/c/d/redis-sidecar.json") {
			t.Fatalf("Want /a/b/c/d/mysql-sidecar.json or /a/b/c/d/redis-sidecar.json but got %d\n", len(m[key]))
		}
	}

	key = "g/h"
	arr = m[key]
	if len(arr) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(arr))
	}
	if arr[0] != "/e/f/g/h/redis-sidecar.json" {
		t.Fatalf("Want /e/f/g/h/redis-sidecar.json but got %d\n", arr[0])
	}
}
