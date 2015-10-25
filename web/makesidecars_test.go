package main

import (
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
)

func TestMakeSidecars(t *testing.T) {
	//	k8s := NewBuilder("url", "admin", "admin123", "key", "sekrit", "us-west-1", &locks.NoOpLocker{}, "repo", "repobranch")
	k8s := DefaultBuilder{}
	buildEvent := v1.UserBuildEvent{Team_: "ae6rt", Project_: "somelib", Refs_: []string{"master"}}

	sidecars := k8s.makeSidecarContainers(buildEvent, map[string]v1.Project{
		"ae6rt/somelib": v1.Project{
			Team:        "ae6rt",
			ProjectName: "somelib",
			Descriptor:  v1.ProjectDescriptor{Image: "magic-image"},
			Sidecars: []string{`
{
    "env": [
        {
            "name": "MYSQL_ROOT_PASSWORD", 
            "value": "r00t"
        }
    ], 
    "image": "mysql:5.6", 
    "name": "mysql", 
    "ports": [
        {
            "containerPort": 3306
        }
    ]
}`, `
{
    "image": "rabbitmq:3.5.4", 
    "name": "rabbitmq", 
    "ports": [
        {
            "containerPort": 5672
        }
    ]
}`,
			},
		},
	})

	if len(sidecars) != 2 {
		t.Fatalf("Want 2 but got %v\n", len(sidecars))
	}

	sidecar := sidecars[0]
	if sidecar.Image != "mysql:5.6" {
		t.Fatalf("Want mysql:5.6 but got %v\n", sidecar.Image)
	}
	if sidecar.Name != "mysql" {
		t.Fatalf("Want mysql but got %v\n", sidecar.Name)
	}
	if len(sidecar.Env) != 1 {
		t.Fatalf("Want 1 but got %v\n", len(sidecar.Env))
	}

	sidecar = sidecars[1]
	if sidecar.Image != "rabbitmq:3.5.4" {
		t.Fatalf("Want rabbitmq:3.5.4 but got %v\n", sidecar.Image)
	}
	if sidecar.Name != "rabbitmq" {
		t.Fatalf("Want rabbitmq but got %v\n", sidecar.Name)
	}
	if len(sidecar.Env) != 0 {
		t.Fatalf("Want 0 but got %v\n", sidecar.Env)
	}
}
