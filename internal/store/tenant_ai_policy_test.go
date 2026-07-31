package store

import "testing"

func TestTenantGeminiKeyUsableRequiresValidStatus(t *testing.T) {
	t.Parallel()

	for _, status := range []string{
		GeminiKeyStatusNone,
		GeminiKeyStatusPresent,
		GeminiKeyStatusInvalid,
		GeminiKeyStatusDegraded,
		"",
		"unknown",
	} {
		if tenantGeminiKeyUsable(status) {
			t.Fatalf("status %q must not be usable for tenant runtime", status)
		}
	}
	if !tenantGeminiKeyUsable(GeminiKeyStatusValid) {
		t.Fatal("validated Gemini key must be usable")
	}
}
