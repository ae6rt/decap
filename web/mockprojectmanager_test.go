package main

import "github.com/ae6rt/decap/web/api/v1"

type ProjectManagerBaseMock struct {
}

func (t *ProjectManagerBaseMock) Assemble() (map[string]v1.Project, error) {
	return nil, nil
}

func (t *ProjectManagerBaseMock) Set(map[string]v1.Project) {
}
