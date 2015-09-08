package main

import "testing"

func TestParentPath(t *testing.T) {
	scriptPath := "/a/b/c/d/build.sh"
	parent := parentPath(scriptPath)
	if parent != "/a/b/c/d" {
		t.Fatalf("Want /a/b/c/d but got %s\n", parent)
	}
}

func TestDescriptorPath(t *testing.T) {
	scriptPath := "/a/b/c/d/build.sh"
	dpath := descriptorPath(scriptPath)
	if dpath != "/a/b/c/d/project.json" {
		t.Fatalf("Want /a/b/c/d/project.json but got %s\n", dpath)
	}
}
