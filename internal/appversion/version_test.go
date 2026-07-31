package appversion

import (
	"os"
	"strings"
	"testing"
)

func TestEmbeddedVersionMatchesReleaseVersion(t *testing.T) {
	releaseVersion, err := os.ReadFile("../../VERSION")
	if err != nil {
		t.Fatalf("read root VERSION: %v", err)
	}
	if got, want := Raw(), strings.TrimSpace(string(releaseVersion)); got != want {
		t.Fatalf("embedded version = %q, root VERSION = %q", got, want)
	}
}
