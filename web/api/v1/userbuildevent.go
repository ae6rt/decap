package v1

import "fmt"

func (e UserBuildEvent) Team() string {
	return e.Team_
}

func (e UserBuildEvent) Project() string {
	return e.Project_
}

func (e UserBuildEvent) Key() string {
	return fmt.Sprintf("%s/%s", e.Team_, e.Project_)
}

func (e UserBuildEvent) Ref() string {
	return e.Ref_
}

/* break this to understand why we needed this */
func (e UserBuildEvent) Hash() string {
	//	return fmt.Sprintf("%s/%s", e.Key(), e.Ref())
	panic("ube hash???")
}
