package store

import (
	"strings"
	"testing"
)

func TestExpandHexColor(t *testing.T) {
	v, err := ExpandHexColor("#AbC")
	if err != nil || v != "#aabbcc" {
		t.Fatalf("got %q %v", v, err)
	}
	v, err = ExpandHexColor("#112233")
	if err != nil || v != "#112233" {
		t.Fatalf("got %q %v", v, err)
	}
	if _, err := ExpandHexColor("red"); err == nil {
		t.Fatal("want error")
	}
}

func TestValidateAndNormalizeTokens(t *testing.T) {
	tok, err := ValidateAndNormalizeTokens(DefaultDarkTokens())
	if err != nil {
		t.Fatal(err)
	}
	if tok["primary"] != "#006dff" {
		t.Fatalf("primary %s", tok["primary"])
	}
	bad := DefaultDarkTokens()
	delete(bad, "accent")
	if _, err := ValidateAndNormalizeTokens(bad); err == nil {
		t.Fatal("want missing key error")
	}
}

func TestValidateBranding(t *testing.T) {
	b, err := ValidateAndNormalizeBranding(ThemeBranding{
		BrandName: "Libra Tech Co.,Ltd",
		Subtitle:  "AI · text & voice",
		LogoURL:   "https://example.com/logo.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if b.BrandName != "Libra Tech Co.,Ltd" {
		t.Fatal(b)
	}
	if _, err := ValidateAndNormalizeBranding(ThemeBranding{BrandName: strings.Repeat("x", 81)}); err == nil {
		t.Fatal("want long name error")
	}
	if _, err := ValidateAndNormalizeBranding(ThemeBranding{LogoURL: "ftp://x"}); err == nil {
		t.Fatal("want bad url")
	}
	if _, err := ValidateAndNormalizeBranding(ThemeBranding{LogoURL: "/api/assets/theme/demo/logo.png"}); err != nil {
		t.Fatal(err)
	}
}

func TestContrastRatioBlackWhite(t *testing.T) {
	r, err := ContrastRatio("#000000", "#ffffff")
	if err != nil {
		t.Fatal(err)
	}
	if r < 20 {
		t.Fatalf("ratio %v", r)
	}
	report := EvaluateContrast(DefaultDarkTokens())
	if !report.OK {
		t.Fatalf("dark default should pass: %+v", report)
	}
	// Force fail
	bad := DefaultDarkTokens()
	bad["text"] = "#111111"
	bad["surface"] = "#121212"
	rep := EvaluateContrast(bad)
	if rep.OK {
		t.Fatal("want fail")
	}
}

func TestResolvePublicBranding(t *testing.T) {
	b := ResolvePublicBranding(ThemeBranding{}, "Acme")
	if b.BrandName != "Acme" || b.Subtitle == "" || b.LogoURL == "" {
		t.Fatalf("%+v", b)
	}
}

func TestCSSVarMap(t *testing.T) {
	m := CSSVarMap(DefaultDarkTokens())
	if m["--mj-primary"] == "" || m["--ink"] == "" {
		t.Fatalf("%v", m)
	}
}

func TestNormalizePreset(t *testing.T) {
	p, err := NormalizePreset("")
	if err != nil || p != "dark" {
		t.Fatal(p, err)
	}
	if _, err := NormalizePreset("neon"); err == nil {
		t.Fatal("want err")
	}
}
