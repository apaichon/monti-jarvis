package metering

import (
	"testing"
	"time"
)

func TestRecorderPriceObservedTokens(t *testing.T) {
	r := &Recorder{rate: RateCard{Version: "r1", InputUnitPriceMicros: 2_000_000, OutputUnitPriceMicros: 4_000_000}}
	state, cost := r.price(UsageEvent{InputTokens: 100, OutputTokens: 50})
	if state != StateObserved || cost != 400 {
		t.Fatalf("price = %s/%d, want observed/400", state, cost)
	}
}

func TestRecorderPriceVoiceIsEstimated(t *testing.T) {
	r := &Recorder{rate: RateCard{Version: "r1", AudioSecondPriceMicros: 10_000}}
	state, cost := r.price(UsageEvent{AudioSeconds: 30})
	if state != StateEstimated || cost != 300_000 {
		t.Fatalf("price = %s/%d, want estimated/300000", state, cost)
	}
}

func TestRecorderNoRateIsUnavailable(t *testing.T) {
	r := &Recorder{rate: RateCard{Version: "unconfigured"}}
	state, cost := r.price(UsageEvent{InputTokens: 1, UsageDate: time.Now()})
	if state != StateUnavailable || cost != 0 {
		t.Fatalf("price = %s/%d, want unavailable/0", state, cost)
	}
}

func TestRecorderPartialTextRateIsUnavailable(t *testing.T) {
	r := &Recorder{rate: RateCard{Version: "r1", OutputUnitPriceMicros: 4_000_000}}
	state, cost := r.price(UsageEvent{InputTokens: 100, OutputTokens: 50})
	if state != StateUnavailable || cost != 0 {
		t.Fatalf("price = %s/%d, want unavailable/0", state, cost)
	}
}

func TestRecorderCostSaturates(t *testing.T) {
	r := &Recorder{rate: RateCard{Version: "r1", InputUnitPriceMicros: 9_000_000_000_000_000_000, OutputUnitPriceMicros: 9_000_000_000_000_000_000}}
	state, cost := r.price(UsageEvent{InputTokens: 1_000_000, OutputTokens: 1_000_000})
	if state != StateObserved || cost != int64(^uint64(0)>>1) {
		t.Fatalf("price = %s/%d, want observed/max-int64", state, cost)
	}
}

func TestStableEventIDDoesNotContainInputs(t *testing.T) {
	id := StableEventID("chat", "tenant-secret", "call-secret", "1")
	if id == "" || len(id) < 10 {
		t.Fatalf("unexpected id %q", id)
	}
	for _, secret := range []string{"tenant-secret", "call-secret"} {
		if contains(id, secret) {
			t.Fatalf("id leaks %q: %s", secret, id)
		}
	}
}

func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
