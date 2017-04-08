package v1

import (
	"fmt"
)

// Key returns the project full key, typically to index a build result in backing storage.
func (t Project) Key() string {
	return fmt.Sprintf("%s/%s", t.Team, t.ProjectName)
}
