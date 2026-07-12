package store

import "testing"

func TestNormalizeLocale(t *testing.T) {
	v, err := NormalizeLocale("TH")
	if err != nil || v != "th" {
		t.Fatalf("got %q %v", v, err)
	}
	if _, err := NormalizeLocale("fr"); err == nil {
		t.Fatal("want invalid")
	}
	v, err = NormalizeOptionalLocale("")
	if err != nil || v != "" {
		t.Fatalf("empty optional %q %v", v, err)
	}
}

func TestValidateTimezone(t *testing.T) {
	if err := ValidateTimezone("Asia/Bangkok"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTimezone("Not/AZone"); err == nil {
		t.Fatal("want invalid")
	}
	if err := ValidateTimezone(""); err == nil {
		t.Fatal("want invalid empty")
	}
}
