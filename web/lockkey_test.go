package main

import (
	"testing"
	"encoding/hex"
)

func TestLockKey(t *testing.T) {
	key := lockKey("foo", "bar")
	if key != hex.EncodeToString([]byte("foo/bar")) {
		t.Fatalf("Want 666f6f2f626172 but got %s\n", key)
	}
}
