package clusterutil

import "strings"

func AsLabel(s string) string {
	forbidden := []string{".", "/"}
	t := s
	for _, v := range forbidden {
		t = strings.Replace(t, v, "_", -1)
		t = strings.Replace(t, v, "_", -1)
		t = strings.Replace(t, v, "_", -1)
	}
	return t
}
