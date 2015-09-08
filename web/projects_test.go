package main

import (
	"os"
	"testing"

	"github.com/ae6rt/ziptools"
)

func TestProjects(t *testing.T) {
	dir, err := ziptools.Unzip("buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}
	proj, err := findProjects("file://"+dir, "master")
	os.RemoveAll(dir)

	if err != nil {
		t.Fatal(err)
	}
	if len(proj) != 3 {
		t.Fatal(err)
	}

	foundIt := false
	for _, v := range proj {
		if v.Parent == "ae6rt" && v.Library == "dynamodb-lab" {
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
	dir, err := ziptools.Unzip("buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}

	projects, err = findProjects("file://"+dir, "master")
	os.RemoveAll(dir)

	if err != nil {
		t.Fatal(err)
	}

	if _, present := findProject("ae6rt", "dynamodb-lab"); !present {
		t.Fatalf("Expecting to find ae6rt/dynamodb-lab project but did not\n")
	}

	if _, present := findProject("nope", "nope"); present {
		t.Fatalf("Not expecting to find nope/nope project but did \n")
	}
}
