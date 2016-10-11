package uuid

import id "github.com/twinj/uuid"

// Uuid returns a new type 4 UUID.
func Uuid() string {
	return id.NewV4().String()
}
