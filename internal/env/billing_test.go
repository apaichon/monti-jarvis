package env

import (
	"testing"
	"time"
)

func TestEnvDurationList(t *testing.T) {
	t.Setenv("TEST_BILLING_RETRIES", "1h,6h,24h")
	got := envDurationList("TEST_BILLING_RETRIES", []time.Duration{time.Minute})
	if len(got) != 3 || got[0] != time.Hour || got[2] != 24*time.Hour {
		t.Fatalf("unexpected durations: %v", got)
	}

	t.Setenv("TEST_BILLING_RETRIES", "bad")
	got = envDurationList("TEST_BILLING_RETRIES", []time.Duration{time.Minute})
	if len(got) != 1 || got[0] != time.Minute {
		t.Fatalf("invalid value should use fallback: %v", got)
	}
}
