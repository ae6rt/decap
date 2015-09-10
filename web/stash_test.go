package main

import (
	"encoding/json"
	"testing"
)

func TestStashEvent(t *testing.T) {
	var event StashEvent
	if err := json.Unmarshal([]byte(`{
   "repository":{
      "slug":"somelib",
      "project":{
         "key":"project"
      }
   },
   "refChanges":[
      {
         "refId":"refs/heads/master",
         "fromHash":"2c847c4e9c2421d038fff26ba82bc859ae6ebe20",
         "toHash":"f259e9032cdeb1e28d073e8a79a1fd6f9587f233",
         "type":"UPDATE"
      }
   ]
}`), &event); err != nil {
		Log.Println(err)
		return
	}

	pushEvent := BuildEvent(event)
	if pushEvent.ProjectKey() != "project/somelib" {
		t.Fatalf("Want project/somelib but got %s\n", pushEvent.ProjectKey())
	}

	branches := pushEvent.Branches()
	if len(branches) != 1 {
		t.Fatalf("Want 1 but got %d\n", len(branches))
	}
	if branches[0] != "master" {
		t.Fatalf("Want master but got %s\n", branches[0])
	}
}
