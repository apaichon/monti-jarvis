package store

import "testing"

func TestGeminiSpeakerVoicesCount(t *testing.T) {
	voices := GeminiSpeakerVoices()
	if len(voices) != 30 {
		t.Fatalf("voice count = %d, want 30 AI Studio speakers", len(voices))
	}
	seen := map[string]bool{}
	for _, v := range voices {
		if v.Name == "" || v.Style == "" || v.Label == "" {
			t.Fatalf("incomplete voice: %+v", v)
		}
		if seen[v.Name] {
			t.Fatalf("duplicate speaker %q", v.Name)
		}
		seen[v.Name] = true
		if v.ProviderID != GeminiLiveProviderID {
			t.Fatalf("provider = %q", v.ProviderID)
		}
	}
	if !IsValidGeminiSpeaker("Aoede") || !IsValidGeminiSpeaker("aoede") {
		t.Fatal("Aoede should be valid")
	}
	if IsValidGeminiSpeaker("NotAVoice") {
		t.Fatal("bogus voice should be invalid")
	}
	canon, ok := CanonicalGeminiSpeaker("puck")
	if !ok || canon != "Puck" {
		t.Fatalf("canonical Puck = %q ok=%v", canon, ok)
	}
}
