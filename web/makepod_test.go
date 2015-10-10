package main

import (
	"testing"

	"github.com/ae6rt/decap/web/k8stypes"
)

func TestMakePod(t *testing.T) {
	k8s := NewBuilder("url", "admin", "admin123", "key", "sekrit", "us-west-1", NoOpLocker{}, "repo", "repobranch")
	buildEvent := UserBuildEvent{Team_: "ae6rt", Project_: "somelib", Refs_: []string{"master"}}

	projectMap := map[string]Project{
		"ae6rt/somelib": Project{
			Team:       "ae6rt",
			Project:    "somelib",
			Descriptor: ProjectDescriptor{Image: "magic-image"},
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

	baseContainer := k8s.makeBaseContainer(buildEvent, "uuid", "master", projectMap)
	sidecars := k8s.makeSidecarContainers(buildEvent, projectMap)

	arr := make([]k8stypes.Container, 0)
	arr = append(arr, baseContainer)
	arr = append(arr, sidecars...)

	pod := k8s.makePod(buildEvent, "uuid", "master", arr)

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
	if labels["project"] != projectMap["ae6rt/somelib"].Project {
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
