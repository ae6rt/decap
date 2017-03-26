package v1

import "fmt"

func (e UserBuildEvent) Team() string {
	return e.Team_
}

func (e UserBuildEvent) Project() string {
	return e.Project_
}

func (e UserBuildEvent) Key() string {
	return fmt.Sprintf("%s/%s/%s", e.Team_, e.Project_, e.Ref_)
}

func (e UserBuildEvent) Ref() string {
	return e.Ref_
}
