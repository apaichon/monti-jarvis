package auth

import "testing"

func TestPasswordHashVerify(t *testing.T) {
	hash, err := HashPassword("demo-admin")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(hash, "demo-admin") {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword(hash, "wrong") {
		t.Fatal("expected wrong password to fail")
	}
}