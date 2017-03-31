package main

import (
	"testing"

	k8sapi "k8s.io/client-go/pkg/api/v1"

	"github.com/ae6rt/decap/web/api/v1"
)

func TestMakePod(t *testing.T) {
	builder := DefaultBuilder{buildScripts: BuildScripts{URL: "repo", Branch: "repobranch"}}

	buildEvent := v1.UserBuildEvent{Team: "ae6rt", Project: "somelib", Ref: "master", ID: "uuid"}

	projectMap := map[string]v1.Project{
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
	}

	baseContainer := builder.makeBaseContainer(buildEvent, projectMap)
	sidecars := builder.makeSidecarContainers(buildEvent, projectMap)

	var arr []k8sapi.Container
	arr = append(arr, baseContainer)
	arr = append(arr, sidecars...)

	pod := builder.makePod(buildEvent, arr)

	if pod.ObjectMeta.Name != "uuid" {
		t.Fatalf("Want uuid but got %v\n", pod.ObjectMeta.Name)
	}
	if pod.ObjectMeta.Namespace != "decap" {
		t.Fatalf("Want decap but got %v\n", pod.ObjectMeta.Namespace)
	}

	labels := pod.ObjectMeta.Labels
	if labels["type"] != "decap-build" {
		t.Fatalf("Want decap-build but got %v\n", labels["type"])
	}
	if labels["team"] != projectMap["ae6rt/somelib"].Team {
		t.Fatalf("Want ae6rt but got %v\n", labels["team"])
	}
	if labels["project"] != projectMap["ae6rt/somelib"].ProjectName {
		t.Fatalf("Want somelib but got %v\n", labels["project"])
	}

	if len(pod.Spec.Volumes) != 2 {
		t.Fatalf("Want 2 but got %v\n", len(pod.Spec.Volumes))
	}

	volume := pod.Spec.Volumes[0]
	if volume.Name != "build-scripts" {
		t.Fatalf("Want build-scripts but got %v\n", volume.Name)
	}
	if volume.VolumeSource.GitRepo.Repository != "repo" {
		t.Fatalf("Want repo but got %v\n", volume.VolumeSource.GitRepo.Repository)
	}
	if volume.VolumeSource.GitRepo.Revision != "repobranch" {
		t.Fatalf("Want repobranch but got %v\n", volume.VolumeSource.GitRepo.Revision)
	}

	volume = pod.Spec.Volumes[1]
	if volume.Name != "decap-credentials" {
		t.Fatalf("Want decap-credentials but got %v\n", volume.Name)
	}
	if volume.VolumeSource.Secret.SecretName != "decap-credentials" {
		t.Fatalf("Want repo but got %v\n", volume.VolumeSource.Secret.SecretName)
	}

	if len(pod.Spec.Containers) != 1+len(sidecars) {
		t.Fatalf("Want %d but got %v\n", 1+len(sidecars), len(pod.Spec.Containers))
	}

	if pod.Spec.RestartPolicy != "Never" {
		t.Fatalf("Want Never but got %v\n", pod.Spec.RestartPolicy)
	}

}
