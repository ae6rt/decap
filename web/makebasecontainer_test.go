package main

import (
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
)

type MakeBaseContainerProjectManagerMock struct {
	ProjectManagerBaseMock
	project v1.Project
}

func (t *MakeBaseContainerProjectManagerMock) Get(key string) *v1.Project {
	return &t.project
}

func TestMakeBaseContainer(t *testing.T) {
	var tests = []struct {
		team        string
		projectName string
		ref         string
		ID          string
		project     v1.Project
	}{
		{
			team:        "ae6rt",
			projectName: "somelib",
			ref:         "master",
			ID:          "uuid",
			project: v1.Project{
				Team:        "ae6rt",
				ProjectName: "somelib",
				Descriptor:  v1.ProjectDescriptor{Image: "magic-image"},
				Sidecars:    []string{},
			},
		},
	}

	// todo This test is pretty verbose.  Is there a better way?  msp april 2017
	for testNumber, test := range tests {
		projectManager := &MakeBaseContainerProjectManagerMock{project: test.project}

		builder := DefaultBuildManager{projectManager: projectManager}

		buildEvent := v1.UserBuildEvent{Team: test.team, Project: test.projectName, Ref: test.ref, ID: test.ID}

		baseContainer := builder.makeBaseContainer(buildEvent)

		if baseContainer.Name != "build-server" {
			t.Errorf("Test %d: want build-server but got %v\n", testNumber, baseContainer.Name)
		}

		if baseContainer.Image != test.project.Descriptor.Image {
			t.Errorf("Test %d: want magic-image but got %v\n", testNumber, baseContainer.Image)
		}

		if len(baseContainer.VolumeMounts) != 2 {
			t.Errorf("Test %d: want 2 but got %v\n", testNumber, len(baseContainer.VolumeMounts))
		}

		i := 0
		if baseContainer.VolumeMounts[i].Name != "build-scripts" {
			t.Errorf("Test %d: want build-scripts but got %v\n", testNumber, baseContainer.VolumeMounts[i].Name)
		}
		if baseContainer.VolumeMounts[i].MountPath != "/home/decap/buildscripts" {
			t.Errorf("Test %d: want /home/decap/buildscripts but got %s\n", testNumber, baseContainer.VolumeMounts[i].MountPath)
		}

		i = i + 1
		if baseContainer.VolumeMounts[i].Name != "decap-credentials" {
			t.Errorf("Test %d: want decap-credentials but got %v\n", testNumber, baseContainer.VolumeMounts[i].Name)
		}
		if baseContainer.VolumeMounts[i].MountPath != "/etc/secrets" {
			t.Errorf("Test %d: want /etc/secrets but got %s\n", testNumber, baseContainer.VolumeMounts[i].MountPath)
		}

		if len(baseContainer.Env) != 4 {
			t.Errorf("Test %d: want 4 but got %v\n", testNumber, len(baseContainer.Env))
		}

		i = 0
		if baseContainer.Env[i].Name != "BUILD_ID" {
			t.Errorf("Test %d: want BUILD_ID but got %v\n", testNumber, baseContainer.Env[i].Name)
		}
		if baseContainer.Env[i].Value != test.ID {
			t.Errorf("Test %d: want uuid but got %v\n", testNumber, baseContainer.Env[i].Value)
		}

		i = i + 1
		if baseContainer.Env[i].Name != "PROJECT_KEY" {
			t.Errorf("Test %d: want PROJECT_KEY but got %v\n", testNumber, baseContainer.Env[i].Name)
		}
		if baseContainer.Env[i].Value != buildEvent.ProjectKey() {
			t.Errorf("Test %d: want ae6rt/somelib but got %v\n", testNumber, baseContainer.Env[i].Value)
		}

		i = i + 1
		if baseContainer.Env[i].Name != "BRANCH_TO_BUILD" {
			t.Errorf("Test %d: want BRANCH_TO_BUILD but got %v\n", testNumber, baseContainer.Env[i].Name)
		}
		if baseContainer.Env[i].Value != test.ref {
			t.Errorf("Test %d: want master but got %v\n", testNumber, baseContainer.Env[i].Value)
		}

		i = i + 1
		if baseContainer.Env[i].Name != "BUILD_LOCK_KEY" {
			t.Errorf("Test %d: want BUILD_LOCK_KEY but got %v\n", testNumber, baseContainer.Env[i].Name)
		}
		if baseContainer.Env[i].Value != buildEvent.Lockname() {
			t.Errorf("Test %d: want opaquekey but got %v\n", testNumber, baseContainer.Env[i].Value)
		}
	}
}
