package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/store"
)

func (s *server) getTenantSatisfactionStatistics(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	today := time.Now().Format("2006-01-02")
	filters := store.SatisfactionStatsFilters{
		StartDate: strings.TrimSpace(r.URL.Query().Get("start_date")),
		EndDate:   strings.TrimSpace(r.URL.Query().Get("end_date")),
		AvatarID:  strings.TrimSpace(r.URL.Query().Get("avatar_id")),
		Channel:   strings.TrimSpace(r.URL.Query().Get("channel")),
	}
	if filters.StartDate == "" {
		filters.StartDate = today
	}
	if filters.EndDate == "" {
		filters.EndDate = filters.StartDate
	}
	if _, err := time.Parse("2006-01-02", filters.StartDate); err != nil {
		writeError(w, http.StatusBadRequest, "start_date must use YYYY-MM-DD")
		return
	}
	if _, err := time.Parse("2006-01-02", filters.EndDate); err != nil {
		writeError(w, http.StatusBadRequest, "end_date must use YYYY-MM-DD")
		return
	}
	if filters.StartDate > filters.EndDate {
		writeError(w, http.StatusBadRequest, "start_date must be before or equal to end_date")
		return
	}
	if filters.Channel != "" && filters.Channel != "chat" && filters.Channel != "voice" {
		writeError(w, http.StatusBadRequest, "channel must be chat or voice")
		return
	}
	stats, err := s.store.GetSatisfactionStatistics(r.Context(), tenantID, filters)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
