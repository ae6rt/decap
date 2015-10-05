package main

import (
	"testing"
)

func TestDescriptorRegex(t *testing.T) {
	var descriptor AtomDescriptor

	descriptor, _ = descriptorForTeamProject([]byte(`{
     "build-image": "ae6rt/java7:latest",
     "managed-branch-regex": ".*",
     "repo-manager": "github",
     "repo-url": "https://github.com/ae6rt/hello-world-java.git",
     "repo-description": "Hello world in Java"}`))

	if !descriptor.isManagedBranch("master") {
		t.Fatalf("Want true")
	}

	descriptor, _ = descriptorForTeamProject([]byte(`{
     "build-image": "ae6rt/java7:latest",
     "repo-manager": "github",
     "repo-url": "https://github.com/ae6rt/hello-world-java.git",
     "repo-description": "Hello world in Java"}`))

	if !descriptor.isManagedBranch("master") {
		t.Fatalf("Want true")
	}

	descriptor, _ = descriptorForTeamProject([]byte(`{
     "build-image": "ae6rt/java7:latest",
     "repo-manager": "github",
     "managed-branch-regex": "issue/.*",
     "repo-url": "https://github.com/ae6rt/hello-world-java.git",
     "repo-description": "Hello world in Java"}`))

	if descriptor.isManagedBranch("master") {
		t.Fatalf("Want false")
	}

	descriptor, _ = descriptorForTeamProject([]byte(`{
     "build-image": "ae6rt/java7:latest",
     "repo-manager": "github",
     "managed-branch-regex": "feature/.*",
     "repo-url": "https://github.com/ae6rt/hello-world-java.git",
     "repo-description": "Hello world in Java"}`))

	if !descriptor.isManagedBranch("feature/PLAT-99") {
		t.Fatalf("Want true")
	}
}
