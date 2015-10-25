package main

import (
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/locks"
)

func TestMakeBaseContainer(t *testing.T) {
	builder := DefaultBuilder{
		AWSAccessKeyID:  "key",
		AWSAccessSecret: "sekrit",
		AWSRegion:       "us-west-1",
		Locker:          &locks.NoOpLocker{},
	}

	buildEvent := v1.UserBuildEvent{Team_: "ae6rt", Project_: "somelib", Refs_: []string{"master"}}
	baseContainer := builder.makeBaseContainer(buildEvent, "uuid", "master", map[string]v1.Project{
		"ae6rt/somelib": v1.Project{Team: "ae6rt", ProjectName: "somelib", Descriptor: v1.ProjectDescriptor{Image: "magic-image"}, Sidecars: []string{}},
	})

	if baseContainer.Name != "build-server" {
		t.Fatalf("Want build-server but got %v\n", baseContainer.Name)
	}

	if baseContainer.Image != "magic-image" {
		t.Fatalf("Want magic-image but got %v\n", baseContainer.Image)
	}

	if len(baseContainer.VolumeMounts) != 2 {
		t.Fatalf("Want 2 but got %v\n", len(baseContainer.VolumeMounts))
	}

	i := 0
	if baseContainer.VolumeMounts[i].Name != "build-scripts" {
		t.Fatalf("Want build-scripts but got %v\n", baseContainer.VolumeMounts[i].Name)
	}
	if baseContainer.VolumeMounts[i].MountPath != "/home/decap/buildscripts" {
		t.Fatalf("Want /home/decap/buildscripts but got %s\n", baseContainer.VolumeMounts[i].MountPath)
	}

	i = i + 1
	if baseContainer.VolumeMounts[i].Name != "decap-credentials" {
		t.Fatalf("Want decap-credentials but got %v\n", baseContainer.VolumeMounts[i].Name)
	}
	if baseContainer.VolumeMounts[i].MountPath != "/etc/secrets" {
		t.Fatalf("Want /etc/secrets but got %s\n", baseContainer.VolumeMounts[i].MountPath)
	}

	if len(baseContainer.Env) != 7 {
		t.Fatalf("Want 7 but got %v\n", len(baseContainer.Env))
	}

	i = 0
	if baseContainer.Env[i].Name != "BUILD_ID" {
		t.Fatalf("Want BUILD_ID but got %v\n", baseContainer.Env[i].Name)
	}
	if baseContainer.Env[i].Value != "uuid" {
		t.Fatalf("Want uuid but got %v\n", baseContainer.Env[i].Value)
	}

	i = i + 1
	if baseContainer.Env[i].Name != "PROJECT_KEY" {
		t.Fatalf("Want PROJECT_KEY but got %v\n", baseContainer.Env[i].Name)
	}
	if baseContainer.Env[i].Value != "ae6rt/somelib" {
		t.Fatalf("Want ae6rt/somelib but got %v\n", baseContainer.Env[i].Value)
	}

	i = i + 1
	if baseContainer.Env[i].Name != "BRANCH_TO_BUILD" {
		t.Fatalf("Want BRANCH_TO_BUILD but got %v\n", baseContainer.Env[i].Name)
	}
	if baseContainer.Env[i].Value != "master" {
		t.Fatalf("Want master but got %v\n", baseContainer.Env[i].Value)
	}

	i = i + 1
	if baseContainer.Env[i].Name != "BUILD_LOCK_KEY" {
		t.Fatalf("Want BUILD_LOCK_KEY but got %v\n", baseContainer.Env[i].Name)
	}
	if baseContainer.Env[i].Value != "opaquekey" {
		t.Fatalf("Want opaquekey but got %v\n", baseContainer.Env[i].Value)
	}

	i = i + 1
	if baseContainer.Env[i].Name != "AWS_ACCESS_KEY_ID" {
		t.Fatalf("Want AWS_ACCESS_KEY_ID but got %v\n", baseContainer.Env[i].Name)
	}
	if baseContainer.Env[i].Value != "key" {
		t.Fatalf("Want key but got %v\n", baseContainer.Env[i].Value)
	}

	i = i + 1
	if baseContainer.Env[i].Name != "AWS_SECRET_ACCESS_KEY" {
		t.Fatalf("Want AWS_SECRET_ACCESS_KEY but got %v\n", baseContainer.Env[i].Name)
	}
	if baseContainer.Env[i].Value != "sekrit" {
		t.Fatalf("Want sekrit but got %v\n", baseContainer.Env[i].Value)
	}

	i = i + 1
	if baseContainer.Env[i].Name != "AWS_DEFAULT_REGION" {
		t.Fatalf("Want AWS_DEFAULT_REGION but got %v\n", baseContainer.Env[i].Name)
	}
	if baseContainer.Env[i].Value != "us-west-1" {
		t.Fatalf("Want us-west-1 but got %v\n", baseContainer.Env[i].Value)
	}

	stopAction := baseContainer.Lifecycle.PreStop.Exec.Command
	if len(stopAction) != 8 {
		t.Fatalf("Want 8 but got %v\n", len(stopAction))
	}

	i = 0
	if stopAction[i] != "bctool" {
		t.Fatalf("Want bctoolt but got %v\n", stopAction[i])
	}

	i = i + 1
	if stopAction[i] != "unlock" {
		t.Fatalf("Want unlock but got %v\n", stopAction[i])
	}

	i = i + 1
	if stopAction[i] != "--lockservice-base-url" {
		t.Fatalf("Want --lockservice-base-url but got %v\n", stopAction[i])
	}

	i = i + 1
	if stopAction[i] != "http://lockservice.decap-system:2379" {
		t.Fatalf("Want http://lockservice.decap-system:2379 but got %v\n", stopAction[i])
	}

	i = i + 1
	if stopAction[i] != "--build-id" {
		t.Fatalf("Want --build-id but got %v\n", stopAction[i])
	}

	i = i + 1
	if stopAction[i] != "uuid" {
		t.Fatalf("Want uuid but got %v\n", stopAction[i])
	}

	i = i + 1
	if stopAction[i] != "--build-lock-key" {
		t.Fatalf("Want --build-lock-key but got %v\n", stopAction[i])
	}

	i = i + 1
	if stopAction[i] != "opaquekey" {
		t.Fatalf("Want opaquekey but got %v\n", stopAction[i])
	}
}
