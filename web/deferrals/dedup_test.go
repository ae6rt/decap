package deferrals

import (
	"sort"
	"testing"
)

func TestSortDeferrals(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		input []Deferral
		want  []Deferral
	}{
		{
			input: []Deferral{
				Deferral{ProjectKey: "a", Branch: "1", UnixTime: 10},
				Deferral{ProjectKey: "b", Branch: "2", UnixTime: 0},
			},
			want: []Deferral{
				Deferral{ProjectKey: "b", Branch: "2"},
				Deferral{ProjectKey: "a", Branch: "1"},
			},
		},
	}

	for testNumber, test := range tests {
		a := make([]Deferral, len(test.input), len(test.input))
		copy(a, test.input)
		sort.Sort(ByTime(a))
		for k, v := range test.want {
			if a[k].Key() != v.Key() {
				t.Fatalf("Test %d, sort.Sort(%+v) got %v, want %+v\n", testNumber, test.input, a[k], v)
			}
		}
	}
}

func TestDedupDeferrals(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		input []Deferral
		want  []Deferral
	}{
		{
			// these get sorted on UnixTime, then uniq'd based on the ProjectKey/Branch tuple
			input: []Deferral{
				Deferral{ProjectKey: "p1", Branch: "feature/p1foo", UnixTime: 10},
				Deferral{ProjectKey: "p1", Branch: "feature/p1foo", UnixTime: 0},
				Deferral{ProjectKey: "p1", Branch: "feature/p1foo", UnixTime: 30},
				Deferral{ProjectKey: "p1", Branch: "feature/p1bar", UnixTime: 50},
				Deferral{ProjectKey: "p2", Branch: "feature/p2foo", UnixTime: 20},
				Deferral{ProjectKey: "p3", Branch: "feature/p3foo", UnixTime: 40},
			},
			want: []Deferral{
				Deferral{ProjectKey: "p1", Branch: "feature/p1foo"},
				Deferral{ProjectKey: "p2", Branch: "feature/p2foo"},
				Deferral{ProjectKey: "p1", Branch: "feature/p1foo"},
				Deferral{ProjectKey: "p3", Branch: "feature/p3foo"},
				Deferral{ProjectKey: "p1", Branch: "feature/p1bar"},
			},
		},
	}

	for testNumber, test := range tests {
		a := make([]Deferral, len(test.input), len(test.input))
		copy(a, test.input)

		sort.Sort(ByTime(a))
		got := dedup(a)
		if len(got) != len(test.want) {
			t.Errorf("Test %d, dedup(%+v) want length %d, got %d\n", testNumber, test.want, len(test.want), len(got))
		}

		for k, v := range test.want {
			if got[k].Key() != v.Key() {
				t.Errorf("Test %d, dedup(%+v) got %s, want %+s\n", testNumber, test.input, got[k].Key(), v.Key())
			}
		}
	}
}
