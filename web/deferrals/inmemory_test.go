package deferrals

import (
	"testing"

	"github.com/ae6rt/decap/web/api/v1"
)

func TestDefer(t *testing.T) {
	var tests = []struct {
		events     []v1.UserBuildEvent
		wantEvents []v1.UserBuildEvent
	}{
		{
			events: []v1.UserBuildEvent{
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
				v1.UserBuildEvent{Team_: "t2", Project_: "p666"},
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
			},
			wantEvents: []v1.UserBuildEvent{
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
				v1.UserBuildEvent{Team_: "t2", Project_: "p666"},
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
			},
		},
	}

	for testNumber, test := range tests {
		d := NewInMemoryDeferralService(nil)
		for _, e := range test.events {
			_ = d.Defer(e)
		}

		got, _ := d.List()

		for k, v := range test.wantEvents {
			if got[k].Lockname() != v.Lockname() {
				t.Errorf("Test %d: want %s, got %s\n", testNumber, v, got[k].Lockname())
			}
		}
	}
}

func TestRemove(t *testing.T) {
	var tests = []struct {
		events     []v1.UserBuildEvent
		removeID   string
		wantEvents []v1.UserBuildEvent
	}{
		{
			events: []v1.UserBuildEvent{
				v1.UserBuildEvent{Team_: "t1", Project_: "p1", ID: "1"},
				v1.UserBuildEvent{Team_: "t1", Project_: "p1", ID: "2"},
				v1.UserBuildEvent{Team_: "t2", Project_: "p666", ID: "3"},
				v1.UserBuildEvent{Team_: "t1", Project_: "p1", ID: "4"},
			},
			removeID: "1",
			wantEvents: []v1.UserBuildEvent{
				v1.UserBuildEvent{Team_: "t2", Project_: "p666", ID: "3"},
				v1.UserBuildEvent{Team_: "t1", Project_: "p1", ID: "4"},
			},
		},
	}

	for testNumber, test := range tests {
		d := NewInMemoryDeferralService(nil)
		for _, e := range test.events {
			_ = d.Defer(e)
		}

		_ = d.Remove(test.removeID)

		got, _ := d.List()

		for k, v := range test.wantEvents {
			if got[k].Lockname() != v.Lockname() {
				t.Errorf("Test %d: want %s, got %s\n", testNumber, v, got[k].Lockname())
			}
		}
	}
}

func TestPoll(t *testing.T) {
	var tests = []struct {
		events     []v1.UserBuildEvent
		wantEvents []v1.UserBuildEvent
	}{
		{
			events: []v1.UserBuildEvent{
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
				v1.UserBuildEvent{Team_: "t2", Project_: "p666"},
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
			},
			wantEvents: []v1.UserBuildEvent{
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
				v1.UserBuildEvent{Team_: "t2", Project_: "p666"},
				v1.UserBuildEvent{Team_: "t1", Project_: "p1"},
			},
		},
	}

	for testNumber, test := range tests {
		d := NewInMemoryDeferralService(nil)
		for _, e := range test.events {
			_ = d.Defer(e)
		}

		got, _ := d.Poll()

		for k, v := range test.wantEvents {
			if got[k].Lockname() != v.Lockname() {
				t.Errorf("Test %d: want %s, got %s\n", testNumber, v, got[k].Lockname())
			}
		}

		l, _ := d.List()
		if len(l) != 0 {
			t.Errorf("Test %d: list should be empty after a poll: %d\n", testNumber, len(l))
		}
	}
}
