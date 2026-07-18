package clickhouse

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQueryCallCenterStatsAggregatesDimensions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		query := string(body)
		for _, expected := range []string{"tenant_id = 'tenant''s'", "toDate('2026-07-14')", "FINAL"} {
			if !strings.Contains(query, expected) {
				t.Fatalf("query missing %q: %s", expected, query)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[
{"avatar_id":"ava","channel":"voice","sessions":"2","total_duration_seconds":"90","freshness":"2026-07-14 09:00:00"},
{"avatar_id":"ava","channel":"chat","sessions":"1","total_duration_seconds":"30","freshness":"2026-07-14 09:05:00"},
{"avatar_id":"neo","channel":"voice","sessions":"1","total_duration_seconds":"120","freshness":"2026-07-14 09:03:00"}]}`))
	}))
	defer server.Close()

	stats, err := New(server.URL, "monti_jarvis", "", "").QueryCallCenterStats(context.Background(), "tenant's", "2026-07-14", "2026-07-14")
	if err != nil {
		t.Fatal(err)
	}
	if stats.CompletedConversations != 4 || stats.TotalDurationSeconds != 240 {
		t.Fatalf("unexpected totals: %+v", stats)
	}
	if stats.AverageDurationSeconds != 60 {
		t.Fatalf("unexpected average: %v", stats.AverageDurationSeconds)
	}
	if len(stats.ByAvatar) != 2 || len(stats.ByChannel) != 2 {
		t.Fatalf("unexpected dimensions: avatars=%+v channels=%+v", stats.ByAvatar, stats.ByChannel)
	}
	if stats.Freshness.IsZero() || stats.Freshness.Hour() != 9 || stats.Freshness.Minute() != 5 {
		t.Fatalf("unexpected freshness: %v", stats.Freshness)
	}
}

func TestAverageSecondsHandlesEmptyCount(t *testing.T) {
	if averageSeconds(100, 0) != 0 {
		t.Fatal("expected zero average for empty count")
	}
	if averageSeconds(100, 4) != 25 {
		t.Fatal("expected exact average")
	}
}

func TestQueryPlatformCallCenterStatsAggregatesTenants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		query := string(body)
		for _, expected := range []string{
			"GROUP BY tenant_id, avatar_id, channel",
			"status = 'archived'",
			"tenant_id = 'tenant''s'",
			"FINAL",
		} {
			if !strings.Contains(query, expected) {
				t.Fatalf("query missing %q: %s", expected, query)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[
{"tenant_id":"tenant's","avatar_id":"ava","channel":"voice","completed_conversations":"2","total_duration_seconds":"90","freshness":"2026-07-14 09:00:00"},
{"tenant_id":"tenant's","avatar_id":"ava","channel":"chat","completed_conversations":"1","total_duration_seconds":"30","freshness":"2026-07-14 09:05:00"},
{"tenant_id":"tenant-2","avatar_id":"neo","channel":"voice","completed_conversations":"1","total_duration_seconds":"120","freshness":"2026-07-14 09:03:00"}]}`))
	}))
	defer server.Close()

	stats, err := New(server.URL, "monti_jarvis", "", "").QueryPlatformCallCenterStats(context.Background(), "tenant's", "2026-07-14", "2026-07-14")
	if err != nil {
		t.Fatal(err)
	}
	if stats.CompletedConversations != 4 || stats.TotalDurationSeconds != 240 {
		t.Fatalf("unexpected totals: %+v", stats)
	}
	if len(stats.ByTenant) != 2 || stats.ByTenant[0].CompletedConversations != 3 {
		t.Fatalf("unexpected tenant aggregates: %+v", stats.ByTenant)
	}
	if len(stats.ByAvatar) != 2 || len(stats.ByChannel) != 2 {
		t.Fatalf("unexpected dimensions: avatars=%+v channels=%+v", stats.ByAvatar, stats.ByChannel)
	}
}

func TestQueryPlatformCallCenterStatsReturnsSourceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "clickhouse unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := New(server.URL, "monti_jarvis", "", "").QueryPlatformCallCenterStats(context.Background(), "tenant-1", "2026-07-18", "2026-07-18")
	if err == nil {
		t.Fatal("expected ClickHouse source error")
	}
}
