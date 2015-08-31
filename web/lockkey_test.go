package main

import (
	"testing"
	"net/url"
)

func TestLockKey(t *testing.T) {
	key := lockKey("foo", "bar")
	if key != url.QueryEscape("foo/bar") {
		t.Fatalf("Want foo%2Fbar but got %s\n", key)
	}
}
