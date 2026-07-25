package store

import (
	"encoding/json"
	"testing"

	"github.com/libra/monti-jarvis/internal/packages"
)

func TestAiaasSeedPackages(t *testing.T) {
	seeds := aiaasSeedPackages()
	if len(seeds) != 4 {
		t.Fatalf("seed package count = %d, want 4", len(seeds))
	}

	want := map[string]struct {
		price                                   int
		avatar, km, storage, concurrent, mobile int
	}{
		"aiaas-500":  {50000, 1, 100, 5368709120, 1, 100},
		"aiaas-1000": {100000, 3, 300, 21474836480, 2, 300},
		"aiaas-1500": {150000, 5, 750, 53687091200, 5, 750},
		"aiaas-2000": {200000, 10, 1500, 107374182400, 10, 1500},
	}
	seen := make(map[string]bool, len(seeds))
	for _, seed := range seeds {
		if seen[seed.slug] {
			t.Fatalf("duplicate seed slug %q", seed.slug)
		}
		seen[seed.slug] = true
		wantSeed, ok := want[seed.slug]
		if !ok {
			t.Fatalf("unexpected seed slug %q", seed.slug)
		}
		if seed.priceCents != wantSeed.price {
			t.Errorf("%s price = %d, want %d", seed.slug, seed.priceCents, wantSeed.price)
		}
		var rules map[string]any
		if err := json.Unmarshal([]byte(seed.rules), &rules); err != nil {
			t.Fatalf("%s rules: %v", seed.slug, err)
		}
		if err := packages.ValidateRules([]byte(rulesV2Fields), rules); err != nil {
			t.Fatalf("%s rules-v2 validation: %v", seed.slug, err)
		}
		for key, wantValue := range map[string]int{
			"max_ai_employees":        wantSeed.avatar,
			"max_km_documents":        wantSeed.km,
			"max_storage_bytes":       wantSeed.storage,
			"max_concurrent_calls":    wantSeed.concurrent,
			"max_mobile_call_minutes": wantSeed.mobile,
		} {
			if got, ok := rules[key].(float64); !ok || int(got) != wantValue {
				t.Errorf("%s %s = %v, want %d", seed.slug, key, rules[key], wantValue)
			}
		}
	}
}
