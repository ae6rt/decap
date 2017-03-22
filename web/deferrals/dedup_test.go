package deferrals

import (
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
)

func TestDedup(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		input []v1.UserBuildEvent
		want  []v1.UserBuildEvent
	}{
		{
			// these get sorted on UnixTime, then uniq'd based on the ProjectKey/Branch tuple
			input: []v1.UserBuildEvent{
				v1.UserBuildEvent{Team_: "p1", Ref_: "b1", ID: "1", DeferredUnixtime: 0},
				v1.UserBuildEvent{Team_: "p1", Ref_: "b1", ID: "2", DeferredUnixtime: 10},
				v1.UserBuildEvent{Team_: "p2", Ref_: "b1", ID: "3", DeferredUnixtime: 20},
				v1.UserBuildEvent{Team_: "p1", Ref_: "b1", ID: "4", DeferredUnixtime: 30},
				v1.UserBuildEvent{Team_: "p3", Ref_: "b3", ID: "5", DeferredUnixtime: 40},
				v1.UserBuildEvent{Team_: "p1", Ref_: "b2", ID: "6", DeferredUnixtime: 50},
			},
			want: []v1.UserBuildEvent{
				v1.UserBuildEvent{Team_: "p1", Ref_: "b1", ID: "1"},
				v1.UserBuildEvent{Team_: "p2", Ref_: "b1", ID: "3"},
				v1.UserBuildEvent{Team_: "p3", Ref_: "b3", ID: "5"},
			},
		},
	}

	for testNumber, test := range tests {
		a := make([]v1.UserBuildEvent, len(test.input), len(test.input))
		copy(a, test.input)

		got := Dedup(a)
		if len(got) != len(test.want) {
			t.Errorf("Test %d, Dedup(%+v) want length %d, got %d\n", testNumber, got, len(test.want), len(got))
		}

		for k, v := range test.want {
			if got[k].ID != v.ID {
				t.Errorf("Test %d, Dedup() got ID %s, want %+s\n", testNumber, got[k].ID, v.ID)
			}
		}
	}
}
