package deferrals

import (
	"sort"

	"github.com/ae6rt/decap/web/api/v1"
)

// ByTime is used to sort v1.UserBuildEvents by Unix time stamp in increasing order.
type ByTime []v1.UserBuildEvent

func (s ByTime) Len() int {
	return len(s)
}
func (s ByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByTime) Less(i, j int) bool {
	return s[i].DeferredUnixtime < s[j].DeferredUnixtime
}

// ByKey is used to sort v1.UserBuildEvents by key, which is project + branch
type ByKey []v1.UserBuildEvent

func (s ByKey) Len() int {
	return len(s)
}
func (s ByKey) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByKey) Less(i, j int) bool {
	return s[i].Key() < s[j].Key()
}

// Dedup returns unique by-key entries sorted by increasing time.
func Dedup(arr []v1.UserBuildEvent) []v1.UserBuildEvent {
	sort.Sort(ByTime(arr))
	sieve := make(map[string]struct{})

	var results []v1.UserBuildEvent
	for _, v := range arr {
		if _, seen := sieve[v.Key()]; !seen {
			results = append(results, v)
			sieve[v.Key()] = struct{}{}
		}
	}

	return results
}
