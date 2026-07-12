package store

import "testing"

func TestNormalizeSlug(t *testing.T) {
	s, err := NormalizeSlug("VIP Support")
	if err != nil || s != "vip-support" {
		t.Fatalf("got %q %v", s, err)
	}
	if _, err := NormalizeSlug("!!!"); err == nil {
		t.Fatal("want invalid")
	}
	s, err = NormalizeSlug("Standard")
	if err != nil || s != "standard" {
		t.Fatalf("got %q %v", s, err)
	}
}
