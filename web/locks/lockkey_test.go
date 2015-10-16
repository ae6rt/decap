package locks

import (
	"testing"
)

func TestLockKey(t *testing.T) {
	key := EtcdLocker{}.Key("foo", "bar")
	if key != ("666f6f2f626172") {
		t.Fatalf("Want 666f6f2f626172 but got %s\n", key)
	}
}
