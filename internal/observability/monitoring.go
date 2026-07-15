package observability

import (
	"context"
	"sort"
	"sync"
	"time"
)

const (
	StatusOperational = "operational"
	StatusDegraded    = "degraded"
	StatusUnavailable = "unavailable"
	StatusDisabled    = "disabled"
	StatusStale       = "stale"

	AnalyticsCurrent     = "current"
	AnalyticsStale       = "stale"
	AnalyticsUnavailable = "unavailable"
	AnalyticsDisabled    = "disabled"
)

type Probe func(context.Context) (enabled bool, err error)

type Dependency struct {
	Name  string
	Probe Probe
}

type Component struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMS *int64 `json:"latency_ms"`
	CheckedAt string `json:"checked_at"`
}

type Analytics struct {
	Status          string     `json:"status"`
	GeneratedAt     *time.Time `json:"generated_at,omitempty"`
	LastProjectedAt *time.Time `json:"last_projected_at,omitempty"`
}

type Snapshot struct {
	OverallStatus string      `json:"overall_status"`
	CheckedAt     time.Time   `json:"checked_at"`
	Components    []Component `json:"components"`
	Analytics     Analytics   `json:"analytics"`
}

type AnalyticsReader func(context.Context) Analytics

type Service struct {
	dependencies []Dependency
	timeout      time.Duration
	now          func() time.Time
}

func New(dependencies []Dependency, timeout time.Duration) *Service {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	ordered := append([]Dependency(nil), dependencies...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	return &Service{dependencies: ordered, timeout: timeout, now: time.Now}
}

func (s *Service) Snapshot(ctx context.Context, analytics AnalyticsReader) Snapshot {
	snapshotCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	now := s.now().UTC()
	components := make([]Component, len(s.dependencies))
	var wg sync.WaitGroup
	for i, dependency := range s.dependencies {
		i, dependency := i, dependency
		wg.Add(1)
		go func() {
			defer wg.Done()
			components[i] = s.probe(snapshotCtx, dependency, now)
		}()
	}
	wg.Wait()

	snapshot := Snapshot{
		CheckedAt:  now,
		Components: components,
		Analytics:  Analytics{Status: AnalyticsDisabled},
	}
	if analytics != nil {
		snapshot.Analytics = analytics(snapshotCtx)
	}
	snapshot.OverallStatus = overallStatus(components, snapshot.Analytics.Status)
	return snapshot
}

func (s *Service) probe(parent context.Context, dependency Dependency, checkedAt time.Time) Component {
	component := Component{Name: dependency.Name, Status: StatusDisabled, CheckedAt: checkedAt.Format(time.RFC3339)}
	if dependency.Probe == nil {
		return component
	}
	started := time.Now()
	ctx, cancel := context.WithTimeout(parent, s.timeout)
	defer cancel()
	enabled, err := dependency.Probe(ctx)
	if !enabled {
		return component
	}
	if err != nil {
		component.Status = StatusUnavailable
	} else {
		component.Status = StatusOperational
	}
	latency := time.Since(started).Milliseconds()
	if latency < 0 {
		latency = 0
	}
	component.LatencyMS = &latency
	return component
}

func overallStatus(components []Component, analyticsStatus string) string {
	for _, component := range components {
		if component.Status == StatusUnavailable {
			return StatusUnavailable
		}
	}
	if analyticsStatus == AnalyticsUnavailable || analyticsStatus == AnalyticsStale {
		return StatusDegraded
	}
	return StatusOperational
}
