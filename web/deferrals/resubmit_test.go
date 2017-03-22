package deferrals

import (
	"fmt"
	"testing"
)

func TestResubmit(t *testing.T) {
	var tests = []struct {
		projectKey string
		branch     string
		unixtime   int64
	}{
		{
			projectKey: "p1",
			branch:     "b1",
			unixtime:   22,
		},
	}

	for testNumber, test := range tests {
		fmt.Println(testNumber, test)
	}
}
