package main

import (
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	k8sapi "k8s.io/client-go/pkg/api/v1"
)

type MakePodProjectManagerMock struct {
	ProjectManagerBaseMock
	project v1.Project
	url     string
	branch  string
}

func (t *MakePodProjectManagerMock) Get(key string) *v1.Project {
	return &t.project
}

func (t *MakePodProjectManagerMock) RepositoryURL() string {
	return t.url
}

func (t *MakePodProjectManagerMock) RepositoryBranch() string {
	return t.branch
}

func TestMakePod(t *testing.T) {
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
				Sidecars: []string{
					`{"env":[{"name":"MYSQL_ROOT_PASSWORD","value":"r00t"}],"image":"mysql:5.6","name":"mysql","ports":[{"containerPort":3306}]}`,
					`{"image":"rabbitmq:3.5.4","name":"rabbitmq","ports":[{"containerPort":5672}]}`,
				},
			},
		},
	}
	for testNumber, test := range tests {
		projectManager := &MakePodProjectManagerMock{project: test.project, url: "repo", branch: "repobranch"}
		builder := DefaultBuildManager{projectManager: projectManager}
		buildEvent := v1.UserBuildEvent{Team: test.team, Project: test.projectName, Ref: test.ref, ID: test.ID}

		baseContainer := builder.makeBaseContainer(buildEvent)
		sidecars := builder.makeSidecarContainers(buildEvent)

		var arr []k8sapi.Container
		arr = append(arr, baseContainer)
		arr = append(arr, sidecars...)

		pod := builder.makePod(buildEvent, arr)

		if pod.ObjectMeta.Name != test.ID {
			t.Errorf("Test %d: Want uuid but got %v\n", testNumber, pod.ObjectMeta.Name)
		}

		if pod.ObjectMeta.Namespace != "decap" {
			t.Errorf("Test %d: Want decap but got %v\n", testNumber, pod.ObjectMeta.Namespace)
		}

		labels := pod.ObjectMeta.Labels
		if labels["type"] != "decap-build" {
			t.Errorf("Test %d: want decap-build but got %v\n", testNumber, labels["type"])
		}
		if labels["team"] != projectManager.Get(buildEvent.ProjectKey()).Team {
			t.Errorf("Test %d: want %s but got %v\n", testNumber, buildEvent.Team, labels["team"])
		}
		if labels["project"] != projectManager.Get(buildEvent.ProjectKey()).ProjectName {
			t.Errorf("Test %d: want somelib but got %v\n", testNumber, labels["project"])
		}

		if len(pod.Spec.Volumes) != 2 {
			t.Errorf("Test %d: want 2 but got %v\n", testNumber, len(pod.Spec.Volumes))
		}

		volume := pod.Spec.Volumes[0]
		if volume.Name != "build-scripts" {
			t.Errorf("Test %d: want build-scripts but got %v\n", testNumber, volume.Name)
		}
		if volume.VolumeSource.GitRepo.Repository != "repo" {
			t.Errorf("Test %d: want repo but got %v\n", testNumber, volume.VolumeSource.GitRepo.Repository)
		}
		if volume.VolumeSource.GitRepo.Revision != "repobranch" {
			t.Errorf("Test %d: want repobranch but got %v\n", testNumber, volume.VolumeSource.GitRepo.Revision)
		}

		volume = pod.Spec.Volumes[1]
		if volume.Name != "decap-credentials" {
			t.Errorf("Test %d: want decap-credentials but got %v\n", testNumber, volume.Name)
		}
		if volume.VolumeSource.Secret.SecretName != "decap-credentials" {
			t.Errorf("Test %d: want repo but got %v\n", testNumber, volume.VolumeSource.Secret.SecretName)
		}

		if pod.Spec.RestartPolicy != "Never" {
			t.Errorf("Test %d: want Never but got %v\n", testNumber, pod.Spec.RestartPolicy)
		}

		// Test side car assembly
		if len(pod.Spec.Containers) != 1+len(sidecars) {
			t.Errorf("Test %d: want sidecar length %d, got %v\n", testNumber, 1+len(sidecars), len(pod.Spec.Containers))
		}
		for k, v := range []string{"mysql", "rabbitmq"} {
			if pod.Spec.Containers[k+1].Name != v {
				t.Errorf("Test %d: want %s, got %s\n", testNumber, v, pod.Spec.Containers[k+1].Name)
			}
		}
	}
}
