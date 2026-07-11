package quota

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/libra/monti-jarvis/internal/entitlements"
	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/redis/go-redis/v9"
)

type fakeEnts struct {
	eff *entitlements.Effective
	err error
}

func (f *fakeEnts) GetEffective(ctx context.Context, tenantID string) (*entitlements.Effective, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.eff, nil
}

type fakeStore struct {
	kmDocs  int
	avatars int
	kmErr   error
	avErr   error
}

func (f *fakeStore) CountTenantKnowledgeDocuments(ctx context.Context, tenantID string) (int, error) {
	return f.kmDocs, f.kmErr
}

func (f *fakeStore) CountActiveTenantAssignments(ctx context.Context, tenantID string) (int, error) {
	return f.avatars, f.avErr
}

func starterRules() map[string]any {
	return map[string]any{
		"max_ai_employees":          2,
		"max_monthly_call_minutes":  500,
		"max_km_documents":          3,
		"max_concurrent_calls":      2,
		"voice_enabled":             true,
		"rag_enabled":               true,
	}
}

func testSvc(t *testing.T, rules map[string]any, us *fakeStore) (*Service, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	if us == nil {
		us = &fakeStore{}
	}
	ents := &fakeEnts{eff: &entitlements.Effective{
		TenantID: "demo",
		Package:  entitlements.PackageSummary{ID: "pkg-starter", Slug: "starter", Name: "Starter"},
		Status:   "active",
		Rules:    rules,
	}}
	cfg := env.Config{
		RedisPrefix:          "monti_jarvis:",
		QuotaEnabled:         true,
		QuotaFailOpen:        true,
		RateLimitEnabled:     true,
		RateLimitChatPerMin:  3,
		RateLimitKMPerMin:    2,
		RateLimitVoicePerMin: 2,
		QuotaConcurrentTTL:   time.Hour,
	}
	svc := NewWithDeps(ents, us, rdb, cfg)
	svc.now = func() time.Time { return time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC) }
	return svc, mr
}

func TestCheckKMDocument_UnderAtOver(t *testing.T) {
	us := &fakeStore{kmDocs: 2}
	svc, _ := testSvc(t, starterRules(), us)
	ctx := context.Background()

	if err := svc.CheckKMDocument(ctx, "demo"); err != nil {
		t.Fatalf("under limit: %v", err)
	}
	us.kmDocs = 3
	err := svc.CheckKMDocument(ctx, "demo")
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("at limit want ErrLimitExceeded, got %v", err)
	}
	var qe *Error
	if !errors.As(err, &qe) || qe.Dimension != DimMaxKMDocuments {
		t.Fatalf("dimension: %#v", err)
	}
	us.kmDocs = 4
	if err := svc.CheckKMDocument(ctx, "demo"); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("over limit: %v", err)
	}
}

func TestAcquireConcurrent_Release(t *testing.T) {
	svc, _ := testSvc(t, starterRules(), nil)
	ctx := context.Background()

	r1, err := svc.AcquireConcurrent(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := svc.AcquireConcurrent(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.AcquireConcurrent(ctx, "demo")
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("third slot should fail: %v", err)
	}
	r1()
	r3, err := svc.AcquireConcurrent(ctx, "demo")
	if err != nil {
		t.Fatalf("after release: %v", err)
	}
	r2()
	r3()
}

func TestAllowRate(t *testing.T) {
	svc, _ := testSvc(t, starterRules(), nil)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := svc.AllowRate(ctx, "demo", BucketChat); err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
	}
	err := svc.AllowRate(ctx, "demo", BucketChat)
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("want rate limited, got %v", err)
	}
}

func TestCheckFeature(t *testing.T) {
	rules := starterRules()
	rules["voice_enabled"] = false
	svc, _ := testSvc(t, rules, nil)
	ctx := context.Background()
	if err := svc.CheckFeature(ctx, "demo", DimVoiceEnabled); !errors.Is(err, ErrFeatureDisabled) {
		t.Fatalf("voice: %v", err)
	}
	if err := svc.CheckFeature(ctx, "demo", DimRAGEnabled); err != nil {
		t.Fatalf("rag should be on: %v", err)
	}
}

func TestCheckAIEmployees(t *testing.T) {
	svc, _ := testSvc(t, starterRules(), &fakeStore{avatars: 2})
	ctx := context.Background()
	if err := svc.CheckAIEmployees(ctx, "demo", 2); err != nil {
		t.Fatalf("at capacity ok for nextCount==limit: %v", err)
	}
	// nextCount 2 with limit 2 is OK (exactly at limit means 2 assigned). Adding 3rd is nextCount=3.
	if err := svc.CheckAIEmployees(ctx, "demo", 3); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("want exceed: %v", err)
	}
}

func TestAddCallMinutesAndMonthlyCheck(t *testing.T) {
	rules := starterRules()
	rules["max_monthly_call_minutes"] = 10
	svc, _ := testSvc(t, rules, nil)
	ctx := context.Background()

	if err := svc.CheckMonthlyMinutes(ctx, "demo", 0); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddCallMinutes(ctx, "demo", 10); err != nil {
		t.Fatal(err)
	}
	if err := svc.CheckMonthlyMinutes(ctx, "demo", 0); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("at limit: %v", err)
	}
	// New month key
	svc.now = func() time.Time { return time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC) }
	if err := svc.CheckMonthlyMinutes(ctx, "demo", 0); err != nil {
		t.Fatalf("new month should reset: %v", err)
	}
}

func TestSnapshot(t *testing.T) {
	us := &fakeStore{kmDocs: 1, avatars: 1}
	svc, _ := testSvc(t, starterRules(), us)
	ctx := context.Background()
	_ = svc.AddCallMinutes(ctx, "demo", 5)
	rel, _ := svc.AcquireConcurrent(ctx, "demo")
	defer rel()

	snap, err := svc.Snapshot(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Status != "active" || snap.Package == nil || snap.Package.Slug != "starter" {
		t.Fatalf("snap package: %+v", snap)
	}
	if snap.Period != "2026-07" {
		t.Fatalf("period %s", snap.Period)
	}
	if snap.Limits == nil || snap.Limits.MaxKMDocuments != 3 {
		t.Fatalf("limits %+v", snap.Limits)
	}
	if snap.Usage.KMDocuments != 1 || snap.Usage.AIEmployees != 1 {
		t.Fatalf("usage %+v", snap.Usage)
	}
	if snap.Usage.MonthlyCallMinutes != 5 || snap.Usage.ConcurrentCalls != 1 {
		t.Fatalf("redis usage %+v", snap.Usage)
	}
}

func TestNoEntitlementFailOpen(t *testing.T) {
	mr, _ := miniredis.Run()
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := env.Config{
		RedisPrefix: "monti_jarvis:", QuotaEnabled: true, QuotaFailOpen: true,
		RateLimitEnabled: true, RateLimitChatPerMin: 10, QuotaConcurrentTTL: time.Hour,
	}
	svc := NewWithDeps(&fakeEnts{err: store.ErrEntitlementNotFound}, &fakeStore{}, rdb, cfg)
	if err := svc.CheckKMDocument(context.Background(), "x"); err != nil {
		t.Fatalf("fail-open: %v", err)
	}
}

func TestNoEntitlementFailClosed(t *testing.T) {
	mr, _ := miniredis.Run()
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := env.Config{
		RedisPrefix: "monti_jarvis:", QuotaEnabled: true, QuotaFailOpen: false,
		QuotaConcurrentTTL: time.Hour,
	}
	svc := NewWithDeps(&fakeEnts{err: store.ErrEntitlementNotFound}, &fakeStore{}, rdb, cfg)
	if err := svc.CheckKMDocument(context.Background(), "x"); !errors.Is(err, ErrNoEntitlement) {
		t.Fatalf("fail-closed: %v", err)
	}
}

func TestStatus(t *testing.T) {
	svc, _ := testSvc(t, starterRules(), nil)
	if got := svc.Status(context.Background()); got != "ok" {
		t.Fatalf("status %s", got)
	}
	if got := svc.RateLimitStatus(context.Background()); got != "ok" {
		t.Fatalf("rl %s", got)
	}
	disabled := NewWithDeps(nil, nil, nil, env.Config{QuotaEnabled: false})
	if disabled.Status(context.Background()) != "disabled" {
		t.Fatal("expected disabled")
	}
}

func TestDisabledSkipsChecks(t *testing.T) {
	svc, _ := testSvc(t, starterRules(), &fakeStore{kmDocs: 999})
	svc.enabled = false
	if err := svc.CheckKMDocument(context.Background(), "demo"); err != nil {
		t.Fatal(err)
	}
}
