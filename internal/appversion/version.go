package appversion

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var raw string

// Raw returns the trimmed VERSION file contents (e.g. "2.24.0").
func Raw() string {
	return strings.TrimSpace(raw)
}

// Display returns the canonical UI/tag form with a leading "v" (e.g. "v2.24.0").
func Display() string {
	v := Raw()
	if v == "" {
		return "v0.0.0-dev"
	}
	if strings.HasPrefix(v, "v") || strings.HasPrefix(v, "V") {
		return "v" + strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V")
	}
	return "v" + v
}
