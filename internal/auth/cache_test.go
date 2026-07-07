package auth

import (
	"context"
	"testing"
)

func TestStripPasswordHashRemovesHash(t *testing.T) {
	c := NewCache(nil, "test:", 0, 0, false)
	user := CachedUser{ID: "u1", Email: "a@b.local", PasswordHash: "hash"}
	if err := c.StripPasswordHash(context.Background(), user); err != nil {
		t.Fatal(err)
	}
}