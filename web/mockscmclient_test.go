package main

type MockScmClient struct {
	team       string
	repository string
	branches   []Ref
	err        error
}

func (c *MockScmClient) GetRefs(team, repo string) ([]Ref, error) {
	c.team = team
	c.repository = repo
	return c.branches, c.err
}
