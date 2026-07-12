package store

import "testing"

func TestNormalizeQuestionHash(t *testing.T) {
	a := NormalizeQuestionHash("  Hello World  ")
	b := NormalizeQuestionHash("hello world")
	if a == "" || a != b {
		t.Fatalf("hash mismatch %q vs %q", a, b)
	}
	c := NormalizeQuestionHash("other")
	if c == a {
		t.Fatal("different questions must hash differently")
	}
}
