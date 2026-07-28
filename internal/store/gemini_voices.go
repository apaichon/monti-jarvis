package store

import "strings"

// GeminiSpeakerVoice is a prebuilt Gemini TTS / Live speaker from
// Google AI Studio generate-speech (https://aistudio.google.com/generate-speech).
// Official list: https://ai.google.dev/gemini-api/docs/speech-generation#voices
type GeminiSpeakerVoice struct {
	Name        string `json:"name"`
	Style       string `json:"style"`
	Label       string `json:"label"`
	ProviderID  string `json:"voice_provider_id"`
	VoiceID     string `json:"voice_id"`
}

const (
	GeminiLiveProviderID = "voice-gemini-live"
	GeminiLiveModelKey   = "gemini-2.5-flash-native-audio-latest"
)

// GeminiSpeakerVoices returns the 30 AI Studio speaker setting names with styles.
func GeminiSpeakerVoices() []GeminiSpeakerVoice {
	type pair struct{ name, style string }
	pairs := []pair{
		{"Zephyr", "Bright"},
		{"Puck", "Upbeat"},
		{"Charon", "Informative"},
		{"Kore", "Firm"},
		{"Fenrir", "Excitable"},
		{"Leda", "Youthful"},
		{"Orus", "Firm"},
		{"Aoede", "Breezy"},
		{"Callirrhoe", "Easy-going"},
		{"Autonoe", "Bright"},
		{"Enceladus", "Breathy"},
		{"Iapetus", "Clear"},
		{"Umbriel", "Easy-going"},
		{"Algieba", "Smooth"},
		{"Despina", "Smooth"},
		{"Erinome", "Clear"},
		{"Algenib", "Gravelly"},
		{"Rasalgethi", "Informative"},
		{"Laomedeia", "Upbeat"},
		{"Achernar", "Soft"},
		{"Alnilam", "Firm"},
		{"Schedar", "Even"},
		{"Gacrux", "Mature"},
		{"Pulcherrima", "Forward"},
		{"Achird", "Friendly"},
		{"Zubenelgenubi", "Casual"},
		{"Vindemiatrix", "Gentle"},
		{"Sadachbia", "Lively"},
		{"Sadaltager", "Knowledgeable"},
		{"Sulafat", "Warm"},
	}
	out := make([]GeminiSpeakerVoice, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, GeminiSpeakerVoice{
			Name:       p.name,
			Style:      p.style,
			Label:      p.name + " — " + p.style,
			ProviderID: GeminiLiveProviderID,
			VoiceID:    GeminiLiveModelKey,
		})
	}
	return out
}

// IsValidGeminiSpeaker returns true when name matches an AI Studio speaker setting.
func IsValidGeminiSpeaker(name string) bool {
	name = strings.TrimSpace(name)
	for _, v := range GeminiSpeakerVoices() {
		if strings.EqualFold(v.Name, name) {
			return true
		}
	}
	return false
}

// CanonicalGeminiSpeaker returns the official casing for a speaker name.
func CanonicalGeminiSpeaker(name string) (string, bool) {
	name = strings.TrimSpace(name)
	for _, v := range GeminiSpeakerVoices() {
		if strings.EqualFold(v.Name, name) {
			return v.Name, true
		}
	}
	return "", false
}
