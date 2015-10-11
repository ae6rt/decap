package main

import "testing"

func TestGetProjectMap(t *testing.T) {
	projectGetChan = make(chan map[string]Project, 1)

	reference := map[string]Project{
		"ae6rt/p1": Project{
			Team:        "ae6rt",
			ProjectName: "p1",
		},
		"wn0owp/p2": Project{
			Team:        "wn0owp",
			ProjectName: "p2",
		},
	}
	projectGetChan <- reference

	dut := getProjects()

	if &reference == &dut {
		t.Fatal("dut is not a copy of reference")
	}
	if len(reference) != len(dut) {
		t.Fatalf("reference size %d and dut size %d projects are not the same size\n", len(reference), len(dut))
	}

	// Spot check Project
	for k, v := range reference {
		if v.Team != dut[k].Team {
			t.Fatalf("reference item %s/%s and dut item %s/%s are not the same\n", k, v.Team, k, dut[k].Team)
		}
		if v.ProjectName != dut[k].ProjectName {
			t.Fatalf("reference item %s/%s and dut item %s/%s are not the same\n", k, v.ProjectName, k, dut[k].ProjectName)
		}
	}
}
