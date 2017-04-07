package clusterutil

import "testing"

func TestAsLabel(t *testing.T) {
	var tests = []struct {
		in     string
		expect string
	}{
		{
			in:     "a",
			expect: "a",
		},
		{
			in:     "a.b",
			expect: "a_b",
		},
		{
			in:     "a/b",
			expect: "a_b",
		},
		{
			in:     "a_b",
			expect: "a_b",
		},
	}
	for testNumber, test := range tests {
		if got := AsLabel(test.in); got != test.expect {
			t.Errorf("Test %d: asLabel(%s) want %s, got %s\n", testNumber, test.in, test.expect, got)
		}
	}
}
