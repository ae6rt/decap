package scmclients

import "github.com/ae6rt/decap/web/api/v1"

type MockScmClient struct {
	Team       string
	Repository string
	Branches   []v1.Ref
	Err        error
}

func (c *MockScmClient) GetRefs(team, repo string) ([]v1.Ref, error) {
	c.Team = team
	c.Repository = repo
	return c.Branches, c.Err
}
