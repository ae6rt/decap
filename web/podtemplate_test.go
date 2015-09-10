package main

import (
	"bytes"
	"testing"

	"text/template"
)

func TestPodJson(t *testing.T) {
	content := []byte(`
{
    "image": "mysql:5.6", 
    "name": "mysql", 
    "ports": [
        {
            "containerPort": 3306
        }
    ]
}`,
	)
	pod := BuildPod{SidecarContainers: content}

	hydratedTemplate := bytes.NewBufferString("")

	theTemplate, err := template.New("test").Parse("Hello json: {{.RawJson .SidecarContainers}}")
	if err != nil {
		t.Fatal(err)
	}
	err = theTemplate.Execute(hydratedTemplate, pod)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Hello json: " + string(content)
	actual := string(hydratedTemplate.Bytes())
	if actual != expected {
		t.Fatalf("Want "+expected+" , but got %s\n", actual)
	}
}
