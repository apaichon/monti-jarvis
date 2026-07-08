package main

import (
	"testing"

	"github.com/libra/monti-jarvis/internal/store"
)

func TestRulesInt(t *testing.T) {
	rules := map[string]any{
		"max_ai_employees": float64(2),
		"missing":          "nope",
	}
	if got, want := rulesInt(rules, "max_ai_employees"), 2; got != want {
		t.Fatalf("rulesInt(max_ai_employees) = %d, want %d", got, want)
	}
	if got, want := rulesInt(rules, "missing"), 0; got != want {
		t.Fatalf("rulesInt(missing) = %d, want %d", got, want)
	}
}

func TestHasActiveVoice(t *testing.T) {
	if hasActiveVoice(nil) {
		t.Fatal("hasActiveVoice(nil) = true, want false")
	}
	voices := []store.AvatarVoice{{Status: "disabled"}, {Status: "active"}}
	if !hasActiveVoice(voices) {
		t.Fatal("hasActiveVoice() = false, want true")
	}
}

func TestBuildAvatarFromBodyValidation(t *testing.T) {
	s := &server{}
	_, _, err := s.buildAvatarFromBody(avatarBody{
		Slug: "Ava",
		Name: "Ava",
		Voices: []avatarVoiceBody{{
			VoiceProviderID: "voice-gemini-live",
			VoiceID:         "model",
			Voice:           "Aoede",
		}},
	}, "", true)
	if err == nil {
		t.Fatal("buildAvatarFromBody() error = nil, want slug lowercase validation error")
	}

	_, _, err = s.buildAvatarFromBody(avatarBody{
		Slug:   "ava",
		Name:   "Ava",
		Status: "active",
		Voices: []avatarVoiceBody{{
			VoiceProviderID: "voice-gemini-live",
			VoiceID:         "model",
			Voice:           "Aoede",
			Status:          "disabled",
		}},
	}, "", true)
	if err == nil {
		t.Fatal("buildAvatarFromBody() error = nil, want active voice validation error")
	}
}