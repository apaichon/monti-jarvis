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

type fakeBonusStore struct {
	balances []store.BonusBalance
	tenant   string
	dim      string
	amount   int64
	key      string
}

func (f *fakeBonusStore) ListTenantBonusBalances(ctx context.Context, tenantID string) ([]store.BonusBalance, error) {
	return f.balances, nil
}

func (f *fakeBonusStore) ConsumeBonus(ctx context.Context, tenantID, dimension string, amount int64, idempotencyKey, sourceType, sourceID string) error {
	f.tenant, f.dim, f.amount, f.key = tenantID, dimension, amount, idempotencyKey
	return nil
}

func (f *fakeStore) CountTenantKnowledgeDocuments(ctx context.Context, tenantID string) (int, error) {
	return f.kmDocs, f.kmErr
}

func (f *fakeStore) CountActiveTenantAssignments(ctx context.Context, tenantID string) (int, error) {
	return f.avatars, f.avErr
}

func starterRules() map[string]any {
	return map[string]any{
		"max_ai_employees":         2,
		"max_monthly_call_minutes": 500,
		"max_km_documents":         3,
		"max_concurrent_calls":     2,
		"voice_enabled":            true,
		"rag_enabled":              true,
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

func TestWaitForQueuedConcurrent_PromotesAfterRelease(t *testing.T) {
	rules := starterRules()
	rules[DimMaxConcurrentCalls] = 1
	svc, _ := testSvc(t, rules, nil)
	svc.callQueueEnabled = true
	svc.callQueueMaxWait = time.Second
	svc.callQueuePositionRefresh = 10 * time.Millisecond
	ctx := context.Background()

	releaseActive, err := svc.AcquireConcurrent(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}

	updates := make(chan QueueUpdate, 4)
	done := make(chan *QueuedAdmission, 1)
	errs := make(chan error, 1)
	go func() {
		admission, err := svc.WaitForQueuedConcurrent(ctx, "demo", "adm-test", func(update QueueUpdate) error {
			updates <- update
			return nil
		})
		if err != nil {
			errs <- err
			return
		}
		done <- admission
	}()

	select {
	case update := <-updates:
		if update.Type != "queue_status" || update.Position != 1 || update.Snapshot.TotalCalls != 2 || update.Snapshot.BusyStatus != "queued" {
			t.Fatalf("queue update = %+v", update)
		}
	case err := <-errs:
		t.Fatalf("wait failed before release: %v", err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for queue status")
	}

	releaseActive()
	select {
	case admission := <-done:
		if admission.AdmissionID != "adm-test" || admission.Release == nil {
			t.Fatalf("admission = %+v", admission)
		}
		if admission.Snapshot.ActiveCalls != 1 || admission.Snapshot.QueuedCallers != 0 || admission.Snapshot.BusyStatus != "admitted" {
			t.Fatalf("admitted snapshot = %+v", admission.Snapshot)
		}
		admission.Release()
	case err := <-errs:
		t.Fatalf("wait failed after release: %v", err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for promotion")
	}
}

func TestWaitForQueuedConcurrent_QueueFull(t *testing.T) {
	rules := starterRules()
	rules[DimMaxConcurrentCalls] = 1
	svc, _ := testSvc(t, rules, nil)
	svc.callQueueEnabled = true
	svc.callQueueMaxPerTenant = 1
	svc.callQueueMaxWait = time.Second
	svc.callQueuePositionRefresh = 10 * time.Millisecond
	ctx := context.Background()

	releaseActive, err := svc.AcquireConcurrent(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	defer releaseActive()

	firstCtx, cancelFirst := context.WithCancel(ctx)
	defer cancelFirst()
	firstReady := make(chan struct{}, 1)
	go func() {
		_, _ = svc.WaitForQueuedConcurrent(firstCtx, "demo", "adm-first", func(update QueueUpdate) error {
			if update.Type == "queue_status" {
				select {
				case firstReady <- struct{}{}:
				default:
				}
			}
			return nil
		})
	}()
	select {
	case <-firstReady:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("first caller did not queue")
	}

	_, err = svc.WaitForQueuedConcurrent(ctx, "demo", "adm-second", nil)
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("second queued caller error = %v, want ErrQueueFull", err)
	}
}

func TestConcurrentQueueSnapshot(t *testing.T) {
	rules := starterRules()
	rules[DimMaxConcurrentCalls] = 1
	svc, _ := testSvc(t, rules, nil)
	svc.callQueueEnabled = true
	svc.callQueueMaxWait = time.Second
	ctx := context.Background()

	releaseActive, err := svc.AcquireConcurrent(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	defer releaseActive()
	if _, err := svc.enqueueCaller(ctx, "demo", "adm-1"); err != nil {
		t.Fatal(err)
	}

	snap, err := svc.ConcurrentQueueSnapshot(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.QueueEnabled || snap.ActiveCalls != 1 || snap.QueuedCallers != 1 || snap.TotalCalls != 2 || snap.MaxConcurrentCalls != 1 || snap.BusyStatus != "queued" {
		t.Fatalf("snapshot = %+v", snap)
	}
}

func TestQueueEntryMetadataAndDuplicateAdmissionID(t *testing.T) {
	rules := starterRules()
	rules[DimMaxConcurrentCalls] = 1
	svc, mr := testSvc(t, rules, nil)
	svc.callQueueEnabled = true
	ctx := context.Background()

	inserted, err := svc.enqueueCaller(ctx, "demo", "adm-dup")
	if err != nil || !inserted {
		t.Fatalf("first enqueue inserted=%v err=%v", inserted, err)
	}
	inserted, err = svc.enqueueCaller(ctx, "demo", "adm-dup")
	if err != nil {
		t.Fatal(err)
	}
	if inserted {
		t.Fatal("duplicate admission id created a second queue position")
	}
	got, err := mr.ZMembers(svc.callQueueKey("demo"))
	if err != nil || len(got) != 1 || got[0] != "adm-dup" {
		t.Fatalf("queue members = %#v err=%v", got, err)
	}
	entryKey := svc.callQueueEntryKey("demo", "adm-dup")
	if mr.HGet(entryKey, "status") != "queued" || mr.HGet(entryKey, "admission_id") != "adm-dup" || mr.HGet(entryKey, "tenant_id") != "demo" || mr.HGet(entryKey, "expires_at_ms") == "" {
		t.Fatalf("entry metadata status=%q admission=%q tenant=%q expires=%q", mr.HGet(entryKey, "status"), mr.HGet(entryKey, "admission_id"), mr.HGet(entryKey, "tenant_id"), mr.HGet(entryKey, "expires_at_ms"))
	}
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

func TestMobileMinutesUseSeparateCounter(t *testing.T) {
	rules := starterRules()
	rules[DimMaxMonthlyCallMinutes] = 10
	rules[DimMaxMobileCallMinutes] = 7
	svc, _ := testSvc(t, rules, nil)
	ctx := context.Background()

	if err := svc.AddCallMinutes(ctx, "demo", 10); err != nil {
		t.Fatal(err)
	}
	if err := svc.CheckMobileMinutes(ctx, "demo", 0); err != nil {
		t.Fatalf("web usage must not consume mobile allowance: %v", err)
	}
	if err := svc.AddMobileCallMinutes(ctx, "demo", 7); err != nil {
		t.Fatal(err)
	}
	if err := svc.CheckMobileMinutes(ctx, "demo", 0); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("mobile allowance should be exhausted: %v", err)
	}
}

func TestLegacyRulesMirrorMonthlyAllowanceForMobile(t *testing.T) {
	limits := limitsFromRules(map[string]any{DimMaxMonthlyCallMinutes: 12})
	if limits.MaxMobileCallMinutes != 12 {
		t.Fatalf("legacy mobile limit = %d, want 12", limits.MaxMobileCallMinutes)
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
	if len(snap.Dimensions) != 6 {
		t.Fatalf("dimension count %d", len(snap.Dimensions))
	}
	for _, row := range snap.Dimensions {
		if row.Consumed == nil || row.Remaining == nil {
			t.Fatalf("available row has nil usage: %+v", row)
		}
		if row.Source == "" || row.Freshness != "current" {
			t.Fatalf("dimension metadata: %+v", row)
		}
	}
}

func TestLimitsWithBonusKeepsBaseAndAddsRemaining(t *testing.T) {
	base := Limits{MaxAIEmployees: 2, MaxMonthlyCallMinutes: 100, MaxKMDocuments: 10, MaxConcurrentCalls: 1}
	total := limitsWithBonus(base, []store.BonusBalance{
		{Dimension: store.BonusAIEmployees, Remaining: 1},
		{Dimension: store.BonusMonthlyCallMinutes, Remaining: 25},
		{Dimension: store.BonusKMDocuments, Remaining: 3},
		{Dimension: store.BonusConcurrentCalls, Remaining: 1},
	})
	if total.MaxAIEmployees != 3 || total.MaxMonthlyCallMinutes != 125 || total.MaxKMDocuments != 13 || total.MaxConcurrentCalls != 2 {
		t.Fatalf("unexpected total limits: %+v", total)
	}
	if base.MaxMonthlyCallMinutes != 100 {
		t.Fatalf("base limits were mutated: %+v", base)
	}
}

func TestConsumeBonusUsageConsumesOnlyNewResourceOverflow(t *testing.T) {
	svc, _ := testSvc(t, starterRules(), &fakeStore{})
	bonus := &fakeBonusStore{balances: []store.BonusBalance{{Dimension: store.BonusKMDocuments, Used: 1, Remaining: 9}}}
	svc.bonus = bonus

	if err := svc.ConsumeBonusUsage(context.Background(), "demo", DimMaxKMDocuments, 6, "km_document", "doc-2"); err != nil {
		t.Fatal(err)
	}
	if bonus.tenant != "demo" || bonus.dim != store.BonusKMDocuments || bonus.amount != 2 {
		t.Fatalf("bonus consume = %s/%s/%d, want demo/km_documents/2", bonus.tenant, bonus.dim, bonus.amount)
	}
	if bonus.key != "resource:km_documents:doc-2:6" {
		t.Fatalf("idempotency key = %q", bonus.key)
	}
}

func TestSnapshotMarksRedisDimensionsUnavailable(t *testing.T) {
	svc, mr := testSvc(t, starterRules(), &fakeStore{kmDocs: 1, avatars: 1})
	mr.Close()
	snap, err := svc.Snapshot(context.Background(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range snap.Dimensions {
		if row.Dimension == "mobile_call_minutes" {
			if row.Source != "unavailable" || row.Freshness != "unavailable" || row.Consumed != nil || row.Remaining != nil {
				t.Fatalf("redis outage row: %+v", row)
			}
			return
		}
	}
	t.Fatal("mobile dimension missing")
}

func TestQuotaUnavailableUsesSafeError(t *testing.T) {
	svc, mr := testSvc(t, starterRules(), nil)
	svc.failOpen = false
	mr.Close()
	err := svc.CheckMobileMinutes(context.Background(), "demo", 0)
	if !errors.Is(err, ErrQuotaUnavailable) {
		t.Fatalf("error = %v, want ErrQuotaUnavailable", err)
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

func TestDailyCallMinutes(t *testing.T) {
	svc, _ := testSvc(t, starterRules(), nil)
	ctx := context.Background()
	tz := "Asia/Bangkok"

	if err := svc.CheckDailyCallMinutes(ctx, "demo", tz, 0); err != nil {
		t.Fatalf("unset daily cap: %v", err)
	}
	if err := svc.CheckDailyCallMinutes(ctx, "demo", tz, 5); err != nil {
		t.Fatalf("fresh day: %v", err)
	}
	if err := svc.AddDailyCallMinutes(ctx, "demo", tz, 5); err != nil {
		t.Fatal(err)
	}
	n, err := svc.GetDailyCallMinutes(ctx, "demo", tz)
	if err != nil || n != 5 {
		t.Fatalf("usage got %d %v", n, err)
	}
	err = svc.CheckDailyCallMinutes(ctx, "demo", tz, 5)
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("at daily cap: %v", err)
	}
	var qe *Error
	if !errors.As(err, &qe) || qe.Code != "daily_call_limit" {
		t.Fatalf("code: %#v", err)
	}
}

func TestDayKeyTimezone(t *testing.T) {
	// 2026-07-11 18:00 UTC = 2026-07-12 01:00 Asia/Bangkok
	now := time.Date(2026, 7, 11, 18, 0, 0, 0, time.UTC)
	if got := dayKey(now, "Asia/Bangkok"); got != "20260712" {
		t.Fatalf("bangkok day %s", got)
	}
	if got := dayKey(now, "UTC"); got != "20260711" {
		t.Fatalf("utc day %s", got)
	}
}

func TestAcquirePreviewConcurrent(t *testing.T) {
	svc, _ := testSvc(t, starterRules(), nil)
	svc.previewMaxConcurrent = 2
	ctx := context.Background()

	r1, err := svc.AcquirePreviewConcurrent(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := svc.AcquirePreviewConcurrent(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.AcquirePreviewConcurrent(ctx, "demo")
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("third: %v", err)
	}
	var qe *Error
	if !errors.As(err, &qe) || qe.Code != "preview_concurrent" {
		t.Fatalf("code %#v", err)
	}
	r1()
	r3, err := svc.AcquirePreviewConcurrent(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	r2()
	r3()
}
