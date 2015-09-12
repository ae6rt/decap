package main

import "testing"

func TestFormatSidecars(t *testing.T) {
	buildPod := BuildPod{}
	r := buildPod.FormatSidecars([]string{"a", "b", "c"})
	if r != ",a,b,c" {
		t.Fatalf("Want ,a,b,c but got %s\n", r)
	}

}
