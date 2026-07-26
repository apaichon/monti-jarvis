package leads

import (
	"net/url"
	"testing"
	"time"
)

func TestValidateLead_OK(t *testing.T) {
	err := ValidateLead(LeadInput{
		Kind: KindContact, Email: "Ops@Example.COM", ConsentContact: true, FullName: "Alex",
	})
	if err != nil {
		t.Fatalf("expected ok: %v", err)
	}
}

func TestValidateLead_ConsentAndHoneypot(t *testing.T) {
	if err := ValidateLead(LeadInput{Kind: KindContact, Email: "a@b.co", ConsentContact: false}); err != ErrConsentRequired {
		t.Fatalf("contact consent: got %v", err)
	}
	if err := ValidateLead(LeadInput{Kind: KindNewsletter, Email: "a@b.co", ConsentMarketing: false}); err != ErrConsentRequired {
		t.Fatalf("newsletter consent: got %v", err)
	}
	if err := ValidateLead(LeadInput{
		Kind: KindContact, Email: "a@b.co", ConsentContact: true, Website: "http://spam",
	}); err != ErrSpam {
		t.Fatalf("honeypot: got %v", err)
	}
	if err := ValidateLead(LeadInput{Kind: "other", Email: "a@b.co", ConsentContact: true}); err != ErrValidation {
		t.Fatalf("kind: got %v", err)
	}
	if err := ValidateLead(LeadInput{Kind: KindContact, Email: "not-an-email", ConsentContact: true}); err != ErrValidation {
		t.Fatalf("email: got %v", err)
	}
}

func TestNormalizeLead_EmailLower(t *testing.T) {
	in := NormalizeLead(LeadInput{Kind: " Contact ", Email: " A@B.Co ", Language: ""})
	if in.Kind != KindContact || in.Email != "a@b.co" || in.Language != "en" {
		t.Fatalf("normalize: %+v", in)
	}
}

func TestValidateFunnel(t *testing.T) {
	if err := ValidateFunnel(FunnelInput{EventName: "page_view", PagePath: "/product/"}); err != nil {
		t.Fatalf("ok: %v", err)
	}
	if err := ValidateFunnel(FunnelInput{EventName: "evil", PagePath: "/product/"}); err != ErrUnknownEvent {
		t.Fatalf("unknown: %v", err)
	}
	if err := ValidateFunnel(FunnelInput{EventName: "cta_click", PagePath: ""}); err != ErrValidation {
		t.Fatalf("path required: %v", err)
	}
}

func TestSafeRelativeRedirect(t *testing.T) {
	got, err := SafeRelativeRedirect("/tenant/register", url.Values{
		"utm_source": {"google"},
		"evil":       {"1"},
		"ref":        {"ABC"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tenant/register?ref=ABC&utm_source=google" && got != "/tenant/register?utm_source=google&ref=ABC" {
		// Encode order is map iteration order — accept either order.
		u, err := url.Parse(got)
		if err != nil {
			t.Fatal(err)
		}
		if u.Path != "/tenant/register" {
			t.Fatalf("path %q", u.Path)
		}
		if u.Query().Get("utm_source") != "google" || u.Query().Get("ref") != "ABC" {
			t.Fatalf("query %v", u.Query())
		}
		if u.Query().Get("evil") != "" {
			t.Fatal("evil query leaked")
		}
	}

	cases := []string{
		"https://evil.com/",
		"//evil.com/x",
		"javascript:alert(1)",
		"data:text/html,hi",
		"/admin/secret",
		"http://localhost/tenant/register",
	}
	for _, c := range cases {
		if _, err := SafeRelativeRedirect(c, nil); err == nil {
			t.Fatalf("expected reject %q", c)
		}
	}

	okPaths := []string{"/", "/product/", "/product/pricing", "/tenant/login", "/tenant/billing"}
	for _, p := range okPaths {
		if _, err := SafeRelativeRedirect(p, nil); err != nil {
			t.Fatalf("path %q: %v", p, err)
		}
	}
}

func TestFilterQuery(t *testing.T) {
	q := FilterQuery(url.Values{
		"utm_campaign": {"aiaas"},
		"password":     {"secret"},
		"lang":         {"th"},
	})
	if q.Get("utm_campaign") != "aiaas" || q.Get("lang") != "th" || q.Get("password") != "" {
		t.Fatalf("filter: %v", q)
	}
}

func TestDedupeWindow(t *testing.T) {
	if !WithinDedupeWindow(time.Now().Add(-time.Hour), 24) {
		t.Fatal("expected inside window")
	}
	if WithinDedupeWindow(time.Now().Add(-48*time.Hour), 24) {
		t.Fatal("expected outside window")
	}
}

func TestProjectPublicPackages(t *testing.T) {
	pkgs := ProjectPublicPackages([]PackageSource{
		{
			ID: "pkg-pro", Name: "Pro", Status: "active", PriceCents: 100000, Currency: "THB",
			BillingPeriod: "monthly",
			Rules: map[string]any{
				"max_ai_employees": 3, "max_km_documents": 300, "voice_enabled": true,
			},
		},
		{ID: "draft", Name: "Draft", Status: "draft", PriceCents: 1},
	})
	if len(pkgs) != 1 {
		t.Fatalf("len=%d", len(pkgs))
	}
	p := pkgs[0]
	if p.PriceAmount != 1000 || p.PriceCurrency != "THB" || p.BillingPeriod != "month" {
		t.Fatalf("price/period: %+v", p)
	}
	if p.RulesSummary["ai_employees"] != 3 {
		t.Fatalf("summary: %+v", p.RulesSummary)
	}
	if len(p.Highlights) == 0 {
		t.Fatal("expected highlights")
	}
	// Ensure no raw internal-only leakage via empty draft exclusion
	for _, pkg := range pkgs {
		if pkg.ID == "draft" {
			t.Fatal("draft should be omitted")
		}
	}
}

func TestValidLeadStatus(t *testing.T) {
	if !ValidLeadStatus("contacted") || ValidLeadStatus("nope") {
		t.Fatal("status set")
	}
}

func TestValidateNoteBody(t *testing.T) {
	if err := ValidateNoteBody("  hi  "); err != nil {
		t.Fatal(err)
	}
	if err := ValidateNoteBody(""); err != ErrNoteEmpty {
		t.Fatal(err)
	}
	if err := ValidateNoteBody(string(make([]byte, 4001))); err != ErrNoteTooLong {
		// make([]byte) fills zeros — still length 4001
		if err != ErrNoteTooLong {
			t.Fatalf("got %v", err)
		}
	}
}
