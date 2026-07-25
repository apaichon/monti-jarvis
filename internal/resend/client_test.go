package resend

import (
	"strings"
	"testing"
)

func TestBrandLogoURL(t *testing.T) {
	if got := BrandLogoURL("https://app.monti.example/"); got != "https://app.monti.example/tenant/images/monti-logo.png" {
		t.Fatalf("BrandLogoURL() = %q", got)
	}
}

func TestCustomerOTPEmailUsesMontiBranding(t *testing.T) {
	subject, body := CustomerOTPEmail("728749703", 600, "https://app.monti.example/tenant/images/monti-logo.png")
	if subject != "Your Monti sign-in code" {
		t.Fatalf("subject = %q", subject)
	}
	for _, want := range []string{"Monti", "728749703", "AI customer operations", "monti-logo.png", "Security reminder"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body does not contain %q", want)
		}
	}
}

func TestVerificationEmailEscapesUserContent(t *testing.T) {
	_, body := VerificationEmail("https://app.monti.example/verify?token=x&next=y", `<Admin>`, BrandLogoURL("https://app.monti.example"))
	if strings.Contains(body, "<Admin>") || !strings.Contains(body, "&lt;Admin&gt;") {
		t.Fatalf("display name was not escaped")
	}
}
