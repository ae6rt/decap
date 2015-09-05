package main

import "github.com/ae6rt/decap/api/v1"

type AWSClient interface {
	GetProjects(pageStart, pageLimit int) (v1.ProjectList, error)
	GetBuilds(project v1.Project, pageStart, pageLimit int) (v1.BuildList, error)
	GetBuild(buildId string) ([]byte, []byte, error) // artifacts, logs, error
}

type DefaultAWSClient struct {
	AWSClient
}

// GetProjects returns one page of Projects
func (c DefaultAWSClient) GetProjects(pageStart, pageLimit int) (v1.ProjectList, error) {
	return v1.ProjectList{}, nil
}

func (c DefaultAWSClient) GetBuilds(project v1.Project, pageStart, pageLimit int) (v1.BuildList, error) {
	return v1.BuildList{}, nil
}

// GetBuild returns the build artifacts as tar gzipped archive, console log as gzip archive, and error
func (c DefaultAWSClient) GetBuild(buildID string) ([]byte, []byte, error) {
	return nil, nil, nil
}
