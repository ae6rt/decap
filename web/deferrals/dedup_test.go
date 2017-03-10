package deferrals

import (
	"testing"
)

func TestDedup(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		input []Deferral
		want  []Deferral
	}{
		{
			input: []Deferral{
				Deferral{ProjectKey: "p1", Branch: "feature/p1foo"},
				Deferral{ProjectKey: "p2", Branch: "feature/p2foo"},
				Deferral{ProjectKey: "p1", Branch: "feature/p1foo"},
				Deferral{ProjectKey: "p3", Branch: "feature/p3foo"},
				Deferral{ProjectKey: "p1", Branch: "feature/p1bar"},
			},
			want: []Deferral{
				Deferral{ProjectKey: "p1", Branch: "feature/p1foo"},
				Deferral{ProjectKey: "p1", Branch: "feature/p1bar"},
				Deferral{ProjectKey: "p2", Branch: "feature/p2foo"},
				Deferral{ProjectKey: "p3", Branch: "feature/p3foo"},
			},
		},
	}

	for testNumber, test := range tests {
		got := dedup(test.input)
		if len(got) != len(test.want) {
			t.Errorf("Test %d, dedup(%+v) want length %d, got %d\n", testNumber, test.want, len(test.want), len(got))
		}

		for _, v := range test.want {
			if !contains(got, v) {
				t.Errorf("Test %d, dedup(%+v) got %+v should contain %+v\n", testNumber, test.input, got, v)
			}
		}
	}

}

func contains(a []Deferral, d Deferral) bool {
	for _, v := range a {
		if v.ProjectKey == d.ProjectKey && v.Branch == d.Branch {
			return true
		}
	}
	return false
}
