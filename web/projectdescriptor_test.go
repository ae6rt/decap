package main

import (
	"testing"
)

func TestDescriptorRegex(t *testing.T) {
	var descriptor AtomDescriptor

	// regex matches all branches
	descriptor, _ = descriptorForTeamProject([]byte(`{
     "build-image": "ae6rt/java7:latest",
     "managed-branch-regex": ".*",
     "repo-manager": "github",
     "repo-url": "https://github.com/ae6rt/hello-world-java.git",
     "repo-description": "Hello world in Java"}`))

	if !descriptor.isRefManaged("master") {
		t.Fatalf("Want true")
	}

	// no regex matches all branches
	descriptor, _ = descriptorForTeamProject([]byte(`{
     "build-image": "ae6rt/java7:latest",
     "repo-manager": "github",
     "repo-url": "https://github.com/ae6rt/hello-world-java.git",
     "repo-description": "Hello world in Java"}`))

	if !descriptor.isRefManaged("master") {
		t.Fatalf("Want true")
	}

	// match only issue/.*
	descriptor, _ = descriptorForTeamProject([]byte(`{
     "build-image": "ae6rt/java7:latest",
     "repo-manager": "github",
     "managed-branch-regex": "issue/.*",
     "repo-url": "https://github.com/ae6rt/hello-world-java.git",
     "repo-description": "Hello world in Java"}`))

	if descriptor.isRefManaged("master") {
		t.Fatalf("Want false")
	}

	// match only feature/.*
	descriptor, _ = descriptorForTeamProject([]byte(`{
     "build-image": "ae6rt/java7:latest",
     "repo-manager": "github",
     "managed-branch-regex": "feature/.*",
     "repo-url": "https://github.com/ae6rt/hello-world-java.git",
     "repo-description": "Hello world in Java"}`))

	if !descriptor.isRefManaged("feature/PLAT-99") {
		t.Fatalf("Want true")
	}
}
