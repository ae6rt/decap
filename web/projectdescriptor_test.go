package main

import (
	"testing"
)

func TestDescriptorRegex(t *testing.T) {
	var descriptor ProjectDescriptor

	// regex matches all branches
	descriptor, _ = descriptorForTeamProject([]byte(`{
     "buildImage": "ae6rt/java7:latest",
     "managedRefRegex": ".*",
     "repoManager": "github",
     "repoUrl": "https://github.com/ae6rt/hello-world-java.git",
     "repoDescription": "Hello world in Java"}`))

	if !descriptor.isRefManaged("master") {
		t.Fatalf("Want true")
	}

	// no regex matches all branches
	descriptor, _ = descriptorForTeamProject([]byte(`{
     "buildImage": "ae6rt/java7:latest",
     "repoManager": "github",
     "repoUrl": "https://github.com/ae6rt/hello-world-java.git",
     "repoDescription": "Hello world in Java"}`))

	if !descriptor.isRefManaged("master") {
		t.Fatalf("Want true")
	}

	// match only issue/.*
	descriptor, _ = descriptorForTeamProject([]byte(`{
     "buildImage": "ae6rt/java7:latest",
     "repoManager": "github",
     "managedRefRegex": "issue/.*",
     "repoUrl": "https://github.com/ae6rt/hello-world-java.git",
     "repoDescription": "Hello world in Java"}`))

	if descriptor.isRefManaged("master") {
		t.Fatalf("Want false")
	}

	// match only feature/.*
	descriptor, _ = descriptorForTeamProject([]byte(`{
     "buildImage": "ae6rt/java7:latest",
     "repoManager": "github",
     "managedRefRegex": "feature/.*",
     "repoUrl": "https://github.com/ae6rt/hello-world-java.git",
     "repoDescription": "Hello world in Java"}`))

	if !descriptor.isRefManaged("feature/PLAT-99") {
		t.Fatalf("Want true")
	}
}
