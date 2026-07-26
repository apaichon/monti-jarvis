package leads

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrValidation       = errors.New("lead validation failed")
	ErrConsentRequired  = errors.New("lead consent required")
	ErrSpam             = errors.New("lead spam detected")
	ErrUnknownEvent     = errors.New("unknown funnel event")
	ErrInvalidStatus    = errors.New("invalid lead status")
	ErrInvalidRedirect  = errors.New("invalid redirect target")
	ErrNoteTooLong      = errors.New("note body too long")
	ErrNoteEmpty        = errors.New("note body is required")
)

const (
	KindContact    = "contact"
	KindBookDemo   = "book_demo"
	KindNewsletter = "newsletter"

	StatusNew             = "new"
	StatusContacted       = "contacted"
	StatusDemoScheduled   = "demo_scheduled"
	StatusDemoCompleted   = "demo_completed"
	StatusQualified       = "qualified"
	StatusRegistered      = "registered"
	StatusKYCPending      = "kyc_pending"
	StatusPackageSelected = "package_selected"
	StatusPaid            = "paid"
	StatusConverted       = "converted"
	StatusLost            = "lost"
	StatusUnsubscribed    = "unsubscribed"
)

var (
	emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

	leadKinds = map[string]struct{}{
		KindContact: {}, KindBookDemo: {}, KindNewsletter: {},
	}

	preferredChannels = map[string]struct{}{
		"": {}, "email": {}, "phone": {}, "line": {}, "other": {},
	}

	languages = map[string]struct{}{
		"en": {}, "th": {},
	}

	leadStatuses = map[string]struct{}{
		StatusNew: {}, StatusContacted: {}, StatusDemoScheduled: {},
		StatusDemoCompleted: {}, StatusQualified: {}, StatusRegistered: {},
		StatusKYCPending: {}, StatusPackageSelected: {}, StatusPaid: {},
		StatusConverted: {}, StatusLost: {}, StatusUnsubscribed: {},
	}

	funnelEvents = map[string]struct{}{
		"page_view": {}, "cta_click": {}, "demo_start": {},
		"lead_submit": {}, "register_start": {},
	}
)

// LeadInput is the public lead capture payload after JSON decode.
type LeadInput struct {
	Kind              string
	Email             string
	FullName          string
	CompanyName       string
	Phone             string
	UseCase           string
	PreferredChannel  string
	Language          string
	ConsentMarketing  bool
	ConsentContact    bool
	UTMSource         string
	UTMMedium         string
	UTMCampaign       string
	UTMContent        string
	UTMTerm           string
	ReferralCode      string
	LandingPath       string
	PackageInterestID string
	Website           string // honeypot — must be empty
}

// FunnelInput is the public funnel beacon payload.
type FunnelInput struct {
	EventName    string
	PagePath     string
	CTAID        string
	SessionKey   string
	UTMSource    string
	UTMMedium    string
	UTMCampaign  string
	UTMContent   string
	UTMTerm      string
	ReferralCode string
}

// NormalizeLead trims and lowercases email/kind and bounds free-text fields.
func NormalizeLead(in LeadInput) LeadInput {
	in.Kind = strings.ToLower(strings.TrimSpace(in.Kind))
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.FullName = trimBound(in.FullName, 160)
	in.CompanyName = trimBound(in.CompanyName, 160)
	in.Phone = trimBound(in.Phone, 40)
	in.UseCase = trimBound(in.UseCase, 2000)
	in.PreferredChannel = strings.ToLower(strings.TrimSpace(in.PreferredChannel))
	in.Language = strings.ToLower(strings.TrimSpace(in.Language))
	if in.Language == "" {
		in.Language = "en"
	}
	in.UTMSource = trimBound(in.UTMSource, 120)
	in.UTMMedium = trimBound(in.UTMMedium, 120)
	in.UTMCampaign = trimBound(in.UTMCampaign, 120)
	in.UTMContent = trimBound(in.UTMContent, 120)
	in.UTMTerm = trimBound(in.UTMTerm, 120)
	in.ReferralCode = trimBound(in.ReferralCode, 64)
	in.LandingPath = trimBound(in.LandingPath, 320)
	in.PackageInterestID = trimBound(in.PackageInterestID, 80)
	in.Website = strings.TrimSpace(in.Website)
	return in
}

// ValidateLead enforces kind, email, honeypot, and consent rules.
func ValidateLead(in LeadInput) error {
	in = NormalizeLead(in)
	if in.Website != "" {
		return ErrSpam
	}
	if _, ok := leadKinds[in.Kind]; !ok {
		return ErrValidation
	}
	if in.Email == "" || !emailPattern.MatchString(in.Email) || len(in.Email) > 320 {
		return ErrValidation
	}
	if _, ok := preferredChannels[in.PreferredChannel]; !ok {
		return ErrValidation
	}
	if _, ok := languages[in.Language]; !ok {
		return ErrValidation
	}
	switch in.Kind {
	case KindNewsletter:
		if !in.ConsentMarketing {
			return ErrConsentRequired
		}
	case KindContact, KindBookDemo:
		if !in.ConsentContact {
			return ErrConsentRequired
		}
	}
	return nil
}

// DedupeKey builds the kind+email identity used for unique storage.
func DedupeKey(kind, email string) string {
	return strings.ToLower(strings.TrimSpace(kind)) + "|" + strings.ToLower(strings.TrimSpace(email))
}

// WithinDedupeWindow reports whether createdAt is still inside the window.
func WithinDedupeWindow(createdAt time.Time, windowHours int) bool {
	if windowHours <= 0 {
		windowHours = 24
	}
	if createdAt.IsZero() {
		return false
	}
	return time.Since(createdAt) <= time.Duration(windowHours)*time.Hour
}

// NormalizeFunnel trims and bounds funnel fields.
func NormalizeFunnel(in FunnelInput) FunnelInput {
	in.EventName = strings.ToLower(strings.TrimSpace(in.EventName))
	in.PagePath = trimBound(in.PagePath, 320)
	in.CTAID = trimBound(in.CTAID, 120)
	in.SessionKey = trimBound(in.SessionKey, 120)
	in.UTMSource = trimBound(in.UTMSource, 120)
	in.UTMMedium = trimBound(in.UTMMedium, 120)
	in.UTMCampaign = trimBound(in.UTMCampaign, 120)
	in.UTMContent = trimBound(in.UTMContent, 120)
	in.UTMTerm = trimBound(in.UTMTerm, 120)
	in.ReferralCode = trimBound(in.ReferralCode, 64)
	return in
}

// ValidateFunnel checks allowlisted event names and basic shape.
func ValidateFunnel(in FunnelInput) error {
	in = NormalizeFunnel(in)
	if _, ok := funnelEvents[in.EventName]; !ok {
		return ErrUnknownEvent
	}
	if in.PagePath == "" {
		return ErrValidation
	}
	return nil
}

// ValidLeadStatus reports whether status is in the sales lifecycle set.
func ValidLeadStatus(status string) bool {
	_, ok := leadStatuses[strings.TrimSpace(strings.ToLower(status))]
	return ok
}

// ValidateNoteBody checks platform note length.
func ValidateNoteBody(body string) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return ErrNoteEmpty
	}
	if len(body) > 4000 {
		return ErrNoteTooLong
	}
	return nil
}

func trimBound(s string, max int) string {
	s = strings.TrimSpace(s)
	if max > 0 && len(s) > max {
		return s[:max]
	}
	return s
}
