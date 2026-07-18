package clickhouse

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQueryAIUsageStatsAggregatesStates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		query := string(body)
		for _, expected := range []string{"ai_cost_usage_facts FINAL", "countIf", "tenant_id = 'tenant''s'", "toDate('2026-07-17')"} {
			if !strings.Contains(query, expected) {
				t.Fatalf("query missing %q: %s", expected, query)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[
{"tenant_id":"tenant's","events":"3","observed_events":"2","estimated_events":"1","unavailable_events":"0","observed_cost_microunits":"1000","estimated_cost_microunits":"200","input_units":"20","output_units":"10","audio_seconds":"3","freshness":"2026-07-17 09:00:00"},
{"tenant_id":"tenant-2","events":"1","observed_events":"0","estimated_events":"0","unavailable_events":"1","observed_cost_microunits":"0","estimated_cost_microunits":"0","input_units":"0","output_units":"0","audio_seconds":"0","freshness":"2026-07-17 09:05:00"}]}`))
	}))
	defer server.Close()

	stats, err := New(server.URL, "monti_jarvis", "", "").QueryAIUsageStats(context.Background(), "tenant's", "2026-07-17", "2026-07-17")
	if err != nil {
		t.Fatal(err)
	}
	if stats.Events != 4 || stats.ObservedEvents != 2 || stats.EstimatedEvents != 1 || stats.UnavailableEvents != 1 {
		t.Fatalf("unexpected counts: %+v", stats)
	}
	if stats.ObservedCostMicros != 1000 || stats.EstimatedCostMicros != 200 || len(stats.ByTenant) != 2 {
		t.Fatalf("unexpected aggregates: %+v", stats)
	}
}
