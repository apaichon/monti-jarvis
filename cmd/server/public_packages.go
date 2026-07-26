package main

import (
	"net/http"

	"github.com/libra/monti-jarvis/internal/leads"
)

func (s *server) publicPackages(w http.ResponseWriter, r *http.Request) {
	if s.store == nil || s.store.Health(r.Context()).Postgres != "ok" {
		writeLeadError(w, http.StatusServiceUnavailable, "package catalog unavailable", "PACKAGE_PUBLIC_UNAVAILABLE")
		return
	}
	pkgs, err := s.store.ListPublicPackages(r.Context())
	if err != nil {
		writeLeadError(w, http.StatusServiceUnavailable, "package catalog unavailable", "PACKAGE_PUBLIC_UNAVAILABLE")
		return
	}
	if pkgs == nil {
		pkgs = []leads.PublicPackage{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"packages": pkgs})
}
