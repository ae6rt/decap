package main

type MockScmClient struct {
	team       string
	repository string
	branches   []Branch
	err        error
}

func (c *MockScmClient) GetBranches(team, repo string) ([]Branch, error) {
	c.team = team
	c.repository = repo
	return c.branches, c.err
}
