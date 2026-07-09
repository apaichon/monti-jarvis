package payment

import "testing"

func TestMaskSecret(t *testing.T) {
	if MaskSecret("") != "" {
		t.Fatal("empty")
	}
	if MaskSecret("ab") != "****" {
		t.Fatal("short")
	}
	if got := MaskSecret("abcdefghij"); got != "****ghij" {
		t.Fatalf("got %q", got)
	}
}

func TestDefaultCallbackURL(t *testing.T) {
	got := DefaultCallbackURL("http://localhost:8091/")
	want := "http://localhost:8091/api/callbacks/chillpay"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}