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
	var results []v1.UserBuildEvent

	sort.Sort(ByKey(arr))
	last := ""
	for _, v := range arr {
		key := v.Key()
		if key != last {
			results = append(results, v)
			last = key
		}
	}

	sort.Sort(ByTime(results))

	return results
}
