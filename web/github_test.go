package main

import (
	"encoding/json"
	"testing"
)

func TestGithubEvent(t *testing.T) {

	var event GithubEvent
	if err := json.Unmarshal([]byte(`
{
    "ref": "refs/heads/master", 
    "repository": {
        "full_name": "ae6rt/dynamodb-lab", 
        "id": 35129377, 
        "name": "dynamodb-lab", 
        "owner": {
            "email": "ae6rt@users.noreply.github.com", 
            "name": "ae6rt"
        }
    }
}
`), &event); err != nil {
		Log.Println(err)
		return
	}

	pushEvent := event.BuildEvent()

	if pushEvent.Team() != "ae6rt" {
		t.Fatalf("Want ae6rt but got %s\n", pushEvent.Team())
	}
	if pushEvent.Project() != "dynamodb-lab" {
		t.Fatalf("Want dynamodb-lab but got %s\n", pushEvent.Project())
	}
	if pushEvent.Key() != "ae6rt/dynamodb-lab/master" {
		t.Fatalf("Want ae6rt/dynamodb-lab/master but got %s\n", pushEvent.Key())
	}

	branch := pushEvent.Ref()
	if branch != "master" {
		t.Fatalf("Want changes but got %s\n", branch)
	}
}
