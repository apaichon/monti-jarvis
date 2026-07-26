package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libra/monti-jarvis/internal/leads"
)

func TestWriteLeadErrorShape(t *testing.T) {
	rec := httptest.NewRecorder()
	writeLeadError(rec, http.StatusTooManyRequests, "too many", "LEAD_RATE_LIMITED")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("code %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "LEAD_RATE_LIMITED" {
		t.Fatalf("body %+v", body)
	}
}

func TestWriteLeadDomainError(t *testing.T) {
	cases := []struct {
		err  error
		code string
		http int
	}{
		{leads.ErrSpam, "LEAD_SPAM", 400},
		{leads.ErrConsentRequired, "LEAD_CONSENT_REQUIRED", 400},
		{leads.ErrValidation, "LEAD_VALIDATION", 400},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		writeLeadDomainError(rec, tc.err)
		if rec.Code != tc.http {
			t.Fatalf("%v status %d", tc.err, rec.Code)
		}
		var body map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&body)
		if body["code"] != tc.code {
			t.Fatalf("%v code %v", tc.err, body["code"])
		}
	}
}

func TestPublicLeadRequestJSONRoundTrip(t *testing.T) {
	raw := `{"kind":"book_demo","email":"ops@example.com","consent_contact":true,"website":"","full_name":"Alex"}`
	var req publicLeadRequest
	if err := json.NewDecoder(bytes.NewReader([]byte(raw))).Decode(&req); err != nil {
		t.Fatal(err)
	}
	in := leads.NormalizeLead(leads.LeadInput{
		Kind: req.Kind, Email: req.Email, ConsentContact: req.ConsentContact,
		FullName: req.FullName, Website: req.Website,
	})
	if err := leads.ValidateLead(in); err != nil {
		t.Fatal(err)
	}
	if in.Email != "ops@example.com" || in.Kind != leads.KindBookDemo {
		t.Fatalf("%+v", in)
	}
}

func TestFunnelEventAllowlistHandlerLogic(t *testing.T) {
	// Unit-level: unknown event maps to FUNNEL_UNKNOWN_EVENT without store.
	if err := leads.ValidateFunnel(leads.FunnelInput{EventName: "x", PagePath: "/p"}); err != leads.ErrUnknownEvent {
		t.Fatal(err)
	}
}

func TestHashClientIPStable(t *testing.T) {
	a := hashClientIP("1.2.3.4")
	b := hashClientIP("1.2.3.4")
	if a == "" || a != b || a == "1.2.3.4" {
		t.Fatalf("hash %q", a)
	}
	if hashClientIP("") != "" {
		t.Fatal("empty")
	}
}
