package projects

import (
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
)

func TestGetProjectMap(t *testing.T) {
	internalAssembly = map[string]v1.Project{
		"ae6rt/p1": v1.Project{
			Team:        "ae6rt",
			ProjectName: "p1",
		},
		"wn0owp/p2": v1.Project{
			Team:        "wn0owp",
			ProjectName: "p2",
		},
	}

	dut := DefaultProjectManager{}.GetProjects()

	if &internalAssembly == &dut {
		t.Fatal("dut is not a copy of reference")
	}
	if len(internalAssembly) != len(dut) {
		t.Errorf("reference size %d and dut size %d projects are not the same size\n", len(internalAssembly), len(dut))
	}

	// Spot check Project
	for k, v := range internalAssembly {
		if v.Team != dut[k].Team {
			t.Errorf("reference item %s/%s and dut item %s/%s are not the same\n", k, v.Team, k, dut[k].Team)
		}
		if v.ProjectName != dut[k].ProjectName {
			t.Errorf("reference item %s/%s and dut item %s/%s are not the same\n", k, v.ProjectName, k, dut[k].ProjectName)
		}
	}
}
