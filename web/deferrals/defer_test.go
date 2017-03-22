package deferrals

import (
	"fmt"
	"testing"
)

func TestDefer(t *testing.T) {
	var tests = []struct {
		wantProjectKey string
		wantBranch     string
		wantBuildID    string
	}{
		{
			wantProjectKey: "proj",
			wantBranch:     "issue/1",
			wantBuildID:    "cf9bc174-96f6-4191-92ee-6355494ebb1e",
		},
	}

	for testNumber, test := range tests {
		fmt.Println(testNumber, test)
	}
}
