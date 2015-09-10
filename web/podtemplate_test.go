package main

import (
	"bytes"
	"testing"

	"text/template"
)

func TestPodJson(t *testing.T) {
	content := `
,{
    "image": "mysql:5.6", 
    "name": "mysql", 
    "ports": [
        {
            "containerPort": 3306
        }
    ]
}`

	pod := BuildPod{SidecarContainers: content}

	hydratedTemplate := bytes.NewBufferString("")
	theTemplate, err := template.New("test").Parse(podTemplate)
	if err != nil {
		t.Fatal(err)
	}
	err = theTemplate.Execute(hydratedTemplate, pod)
	if err != nil {
		t.Fatal(err)
	}

	//	actual := string(hydratedTemplate.Bytes())
	//fmt.Println(actual)
}
