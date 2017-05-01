package projects

import (
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
	kitlog "github.com/go-kit/kit/log"
)

func TestGetProjectMap(t *testing.T) {
	projectsView = map[string]v1.Project{
		"ae6rt/p1": v1.Project{
			Team:        "ae6rt",
			ProjectName: "p1",
		},
		"wn0owp/p2": v1.Project{
			Team:        "wn0owp",
			ProjectName: "p2",
		},
	}

	projectManager := NewDefaultManager("", "", kitlog.NewNopLogger())
	dut := projectManager.GetProjects()

	if &projectsView == &dut {
		t.Fatal("dut is not a copy of reference")
	}
	if len(projectsView) != len(dut) {
		t.Errorf("reference size %d and dut size %d projects are not the same size\n", len(projectsView), len(dut))
	}

	// Spot check Project
	for k, v := range projectsView {
		if v.Team != dut[k].Team {
			t.Errorf("reference item %s/%s and dut item %s/%s are not the same\n", k, v.Team, k, dut[k].Team)
		}
		if v.ProjectName != dut[k].ProjectName {
			t.Errorf("reference item %s/%s and dut item %s/%s are not the same\n", k, v.ProjectName, k, dut[k].ProjectName)
		}
	}
}
