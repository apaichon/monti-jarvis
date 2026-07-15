package observability

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSnapshotNormalizesStatusesAndOrdersComponents(t *testing.T) {
	service := New([]Dependency{
		{Name: "redis", Probe: func(context.Context) (bool, error) { return true, nil }},
		{Name: "postgres", Probe: func(context.Context) (bool, error) { return true, errors.New("secret provider detail") }},
		{Name: "gemini", Probe: func(context.Context) (bool, error) { return false, nil }},
	}, 100*time.Millisecond)

	snapshot := service.Snapshot(context.Background(), func(context.Context) Analytics {
		return Analytics{Status: AnalyticsStale}
	})

	if snapshot.OverallStatus != StatusUnavailable {
		t.Fatalf("overall status = %q, want %q", snapshot.OverallStatus, StatusUnavailable)
	}
	if len(snapshot.Components) != 3 || snapshot.Components[0].Name != "gemini" || snapshot.Components[1].Name != "postgres" || snapshot.Components[2].Name != "redis" {
		t.Fatalf("unexpected component order: %#v", snapshot.Components)
	}
	if snapshot.Components[0].Status != StatusDisabled {
		t.Fatalf("disabled component status = %q", snapshot.Components[0].Status)
	}
	if snapshot.Components[1].Status != StatusUnavailable || snapshot.Components[1].LatencyMS == nil {
		t.Fatalf("failed component = %#v", snapshot.Components[1])
	}
	if snapshot.Components[2].Status != StatusOperational || snapshot.Components[2].LatencyMS == nil {
		t.Fatalf("healthy component = %#v", snapshot.Components[2])
	}
}

func TestSnapshotAnalyticsStaleDegradesHealthyDependencies(t *testing.T) {
	service := New([]Dependency{{Name: "redis", Probe: func(context.Context) (bool, error) { return true, nil }}}, time.Second)
	snapshot := service.Snapshot(context.Background(), func(context.Context) Analytics {
		return Analytics{Status: AnalyticsStale}
	})
	if snapshot.OverallStatus != StatusDegraded {
		t.Fatalf("overall status = %q, want %q", snapshot.OverallStatus, StatusDegraded)
	}
}

func TestSnapshotTimeoutNormalizesDependencyFailure(t *testing.T) {
	service := New([]Dependency{{Name: "redis", Probe: func(ctx context.Context) (bool, error) {
		<-ctx.Done()
		return true, ctx.Err()
	}}}, 10*time.Millisecond)

	started := time.Now()
	snapshot := service.Snapshot(context.Background(), func(ctx context.Context) Analytics {
		if ctx.Err() == nil {
			t.Fatal("analytics context should be canceled after the shared timeout")
		}
		return Analytics{Status: AnalyticsUnavailable}
	})

	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("snapshot took %s, want bounded completion", elapsed)
	}
	if snapshot.Components[0].Status != StatusUnavailable || snapshot.Components[0].LatencyMS == nil {
		t.Fatalf("timed out component = %#v", snapshot.Components[0])
	}
}
