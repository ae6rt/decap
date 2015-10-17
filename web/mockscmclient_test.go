package main

import "github.com/ae6rt/decap/web/api/v1"

type MockScmClient struct {
	team       string
	repository string
	branches   []v1.Ref
	err        error
}

func (c *MockScmClient) GetRefs(team, repo string) ([]v1.Ref, error) {
	c.team = team
	c.repository = repo
	return c.branches, c.err
}
