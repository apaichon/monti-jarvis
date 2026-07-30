package store

import "testing"

func TestAvatarPortraitVariantPaths(t *testing.T) {
	tests := []struct {
		name    string
		variant string
		wantKey string
		wantURL string
	}{
		{
			name:    "default",
			wantKey: "avatars/maya/portrait.webp",
			wantURL: "/api/assets/avatars/maya/portrait.webp",
		},
		{
			name:    "dark",
			variant: "dark",
			wantKey: "avatars/maya/portrait-dark.webp",
			wantURL: "/api/assets/avatars/maya/portrait-dark.webp",
		},
		{
			name:    "light",
			variant: "light",
			wantKey: "avatars/maya/portrait-light.webp",
			wantURL: "/api/assets/avatars/maya/portrait-light.webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AvatarPortraitVariantKey(" Maya ", tt.variant, ".WEBP"); got != tt.wantKey {
				t.Fatalf("AvatarPortraitVariantKey() = %q, want %q", got, tt.wantKey)
			}
			if got := AvatarPortraitVariantURL(" Maya ", tt.variant, ".WEBP"); got != tt.wantURL {
				t.Fatalf("AvatarPortraitVariantURL() = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestValidAvatarPortraitFilename(t *testing.T) {
	for _, name := range []string{"portrait.jpg", "portrait-dark.png", "portrait-light.webp"} {
		if !validAvatarPortraitFilename(name) {
			t.Errorf("validAvatarPortraitFilename(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"avatar.jpg", "dark-portrait.png", "portraitlight.webp"} {
		if validAvatarPortraitFilename(name) {
			t.Errorf("validAvatarPortraitFilename(%q) = true, want false", name)
		}
	}
}
