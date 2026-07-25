package store

import "testing"

func TestNormalizeReferralCode(t *testing.T) {
	if got := normalizeReferralCode("  REF_ABC123 "); got != "ref_abc123" {
		t.Fatalf("normalizeReferralCode() = %q, want %q", got, "ref_abc123")
	}
}

func TestReferralStatuses(t *testing.T) {
	statuses := []string{
		ReferralClicked,
		ReferralAttributed,
		ReferralPending,
		ReferralQualified,
		ReferralRejected,
		ReferralReversed,
	}
	if len(statuses) != 6 {
		t.Fatalf("got %d referral statuses, want 6", len(statuses))
	}
}
