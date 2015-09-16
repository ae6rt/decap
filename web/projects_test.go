package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ae6rt/ziptools"
)

func TestAssembleProjects(t *testing.T) {
	dir, err := ziptools.Unzip("test-data/buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}
	proj, err := assembleProjects("file://"+dir, "master")
	os.RemoveAll(dir)

	if err != nil {
		t.Fatal(err)
	}

	if len(proj) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(proj))
	}

	foundIt := false
	for _, v := range proj {
		if v.Team == "ae6rt" && v.Library == "dynamodb-lab" {
			foundIt = true
			if v.Descriptor.RepoManager != "github" {
				t.Fatalf("Want github but got %s\n", v.Descriptor.RepoManager)
			}
			if v.Descriptor.RepoURL != "https://github.com/ae6rt/dynamodb-lab.git" {
				t.Fatalf("Want https://github.com/ae6rt/dynamodb-lab.git but got %s\n", v.Descriptor.RepoManager)
			}
			break
		}
	}
	if !foundIt {
		t.Fatalf("Want a project ae6rt/library but did not find one\n")
	}
}

func TestProject(t *testing.T) {
	dir, err := ziptools.Unzip("test-data/buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}

	projects, err = assembleProjects("file://"+dir, "master")
	os.RemoveAll(dir)

	if err != nil {
		t.Fatal(err)
	}

	if _, present := projectByTeamLibrary("ae6rt", "dynamodb-lab"); !present {
		t.Fatalf("Expecting to find ae6rt/dynamodb-lab project but did not\n")
	}

	if _, present := projectByTeamLibrary("nope", "nope"); present {
		t.Fatalf("Not expecting to find nope/nope project but did \n")
	}
}

func TestFindBuildScriptsByRegex(t *testing.T) {
	dir, err := ziptools.Unzip("test-data/buildscripts-repo.zip")
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
	for _, v := range files {
		if _, err := ioutil.ReadFile(v); err != nil {
			t.Fatal(err)
		}
	}
	os.RemoveAll(dir)
}

func TestFindProjectDescriptorsByRegex(t *testing.T) {
	dir, err := ziptools.Unzip("test-data/buildscripts-repo.zip")
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
	for _, v := range files {
		if _, err := ioutil.ReadFile(v); err != nil {
			t.Fatal(err)
		}
	}
	os.RemoveAll(dir)
}

func TestFindSidecarsByRegex(t *testing.T) {
	dir, err := ziptools.Unzip("test-data/buildscripts-repo.zip")
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
	for _, v := range files {
		if _, err := ioutil.ReadFile(v); err != nil {
			t.Fatal(err)
		}
	}
	os.RemoveAll(dir)
}

func TestReadSidecars(t *testing.T) {
	dir, err := ziptools.Unzip("test-data/buildscripts-repo.zip")
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
