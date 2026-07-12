package quota

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/libra/monti-jarvis/internal/entitlements"
	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/redis/go-redis/v9"
)

// EntitlementReader loads effective package rules for a tenant.
type EntitlementReader interface {
	GetEffective(ctx context.Context, tenantID string) (*entitlements.Effective, error)
}

// UsageStore provides Postgres-backed usage counts.
type UsageStore interface {
	CountTenantKnowledgeDocuments(ctx context.Context, tenantID string) (int, error)
	CountActiveTenantAssignments(ctx context.Context, tenantID string) (int, error)
}

// Service enforces package quotas and API rate limits (SPRINT-013).
type Service struct {
	ents               EntitlementReader
	store              UsageStore
	rdb                *redis.Client
	prefix             string
	enabled            bool
	rateEnabled        bool
	failOpen           bool
	chatPerMin         int
	kmPerMin           int
	voicePerMin        int
	concurrentTTL      time.Duration
	previewMaxConcurrent int
	now                func() time.Time // injectable for tests
}

// New builds a quota service from app config + store + entitlements.
func New(ents *entitlements.Service, st *store.Store, cfg env.Config) *Service {
	var rdb *redis.Client
	if st != nil {
		rdb = st.Redis()
	}
	var us UsageStore
	if st != nil {
		us = st
	}
	return NewWithDeps(ents, us, rdb, cfg)
}

// NewWithDeps is for tests and alternate wiring.
func NewWithDeps(ents EntitlementReader, us UsageStore, rdb *redis.Client, cfg env.Config) *Service {
	prefix := cfg.RedisPrefix
	if prefix == "" {
		prefix = "monti_jarvis:"
	}
	chat := cfg.RateLimitChatPerMin
	if chat <= 0 {
		chat = 60
	}
	km := cfg.RateLimitKMPerMin
	if km <= 0 {
		km = 30
	}
	voice := cfg.RateLimitVoicePerMin
	if voice <= 0 {
		voice = 20
	}
	ttl := cfg.QuotaConcurrentTTL
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	previewMax := cfg.PreviewMaxConcurrent
	if previewMax <= 0 {
		previewMax = 2
	}
	return &Service{
		ents:                 ents,
		store:                us,
		rdb:                  rdb,
		prefix:               prefix,
		enabled:              cfg.QuotaEnabled && rdb != nil,
		rateEnabled:          cfg.RateLimitEnabled && rdb != nil,
		failOpen:             cfg.QuotaFailOpen,
		chatPerMin:           chat,
		kmPerMin:             km,
		voicePerMin:          voice,
		concurrentTTL:        ttl,
		previewMaxConcurrent: previewMax,
		now:                  time.Now,
	}
}

// Status reports quota health for /api/infra: ok | disabled | degraded.
func (s *Service) Status(ctx context.Context) string {
	if s == nil || !s.enabled {
		return "disabled"
	}
	if s.rdb == nil {
		return "disabled"
	}
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return "degraded"
	}
	return "ok"
}

// RateLimitStatus reports rate-limit health for /api/infra.
func (s *Service) RateLimitStatus(ctx context.Context) string {
	if s == nil || !s.rateEnabled {
		return "disabled"
	}
	if s.rdb == nil {
		return "disabled"
	}
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return "degraded"
	}
	return "ok"
}

// Snapshot returns limits + usage for platform admin UI.
func (s *Service) Snapshot(ctx context.Context, tenantID string) (*Snapshot, error) {
	period := s.now().UTC().Format("2006-01")
	out := &Snapshot{
		TenantID: tenantID,
		Status:   "none",
		Period:   period,
		Usage:    Usage{},
	}
	if s == nil {
		return out, nil
	}

	usage, err := s.collectUsage(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out.Usage = usage

	eff, err := s.effective(ctx, tenantID)
	if err != nil {
		if errors.Is(err, store.ErrEntitlementNotFound) || errors.Is(err, ErrNoEntitlement) {
			return out, nil
		}
		return nil, err
	}
	if eff == nil {
		return out, nil
	}
	limits := limitsFromRules(eff.Rules)
	out.Status = eff.Status
	if out.Status == "" {
		out.Status = "active"
	}
	out.Limits = &limits
	out.Package = &PackageSummary{
		ID:   eff.Package.ID,
		Slug: eff.Package.Slug,
		Name: eff.Package.Name,
	}
	return out, nil
}

// AllowRate enforces per-tenant per-minute rate limits for chat|km|voice.
func (s *Service) AllowRate(ctx context.Context, tenantID, bucket string) error {
	if s == nil || !s.rateEnabled || tenantID == "" {
		return nil
	}
	limit := s.rateLimitFor(bucket)
	if limit <= 0 {
		return nil
	}
	key := s.rateKey(tenantID, bucket)
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return s.onRedisErr("AllowRate", err)
	}
	if n == 1 {
		_ = s.rdb.Expire(ctx, key, 2*time.Minute).Err()
	}
	if int(n) > limit {
		return rateLimited(bucket, limit, int(n))
	}
	return nil
}

// CheckFeature verifies voice_enabled or rag_enabled.
func (s *Service) CheckFeature(ctx context.Context, tenantID, flag string) error {
	if s == nil || !s.enabled {
		return nil
	}
	limits, err := s.limitsOrNil(ctx, tenantID)
	if err != nil {
		return err
	}
	if limits == nil {
		return nil // fail-open path already handled in limitsOrNil
	}
	switch flag {
	case DimVoiceEnabled:
		if !limits.VoiceEnabled {
			return featureDisabled(flag)
		}
	case DimRAGEnabled:
		if !limits.RAGEnabled {
			return featureDisabled(flag)
		}
	default:
		return fmt.Errorf("unknown feature flag %q", flag)
	}
	return nil
}

// CheckKMDocument denies when current doc count already at/over max_km_documents.
func (s *Service) CheckKMDocument(ctx context.Context, tenantID string) error {
	if s == nil || !s.enabled {
		return nil
	}
	limits, err := s.limitsOrNil(ctx, tenantID)
	if err != nil {
		return err
	}
	if limits == nil {
		return nil
	}
	n, err := s.countKM(ctx, tenantID)
	if err != nil {
		return err
	}
	if n >= limits.MaxKMDocuments {
		return limitExceeded(DimMaxKMDocuments, limits.MaxKMDocuments, n)
	}
	return nil
}

// CheckAIEmployees denies when nextCount would exceed max_ai_employees.
func (s *Service) CheckAIEmployees(ctx context.Context, tenantID string, nextCount int) error {
	if s == nil || !s.enabled {
		return nil
	}
	limits, err := s.limitsOrNil(ctx, tenantID)
	if err != nil {
		return err
	}
	if limits == nil {
		return nil
	}
	if nextCount > limits.MaxAIEmployees {
		// usage shown as current assigned = nextCount-1 when adding one
		usage := nextCount - 1
		if usage < 0 {
			usage = 0
		}
		return limitExceeded(DimMaxAIEmployees, limits.MaxAIEmployees, usage)
	}
	return nil
}

// CheckMonthlyMinutes denies when usage would exceed max_monthly_call_minutes.
// additional==0: pre-check before starting a session (deny if already at/over limit).
// additional>0: deny if cur+additional would exceed the limit.
func (s *Service) CheckMonthlyMinutes(ctx context.Context, tenantID string, additional int) error {
	if s == nil || !s.enabled {
		return nil
	}
	if additional < 0 {
		additional = 0
	}
	limits, err := s.limitsOrNil(ctx, tenantID)
	if err != nil {
		return err
	}
	if limits == nil {
		return nil
	}
	cur, err := s.getInt(ctx, s.minutesKey(tenantID))
	if err != nil {
		return s.onRedisErr("CheckMonthlyMinutes", err)
	}
	if additional == 0 {
		if cur >= limits.MaxMonthlyCallMinutes {
			return limitExceeded(DimMaxMonthlyCallMinutes, limits.MaxMonthlyCallMinutes, cur)
		}
		return nil
	}
	if cur+additional > limits.MaxMonthlyCallMinutes {
		return limitExceeded(DimMaxMonthlyCallMinutes, limits.MaxMonthlyCallMinutes, cur)
	}
	return nil
}

// AcquireConcurrent reserves one concurrent call slot. Caller must invoke release.
func (s *Service) AcquireConcurrent(ctx context.Context, tenantID string) (release func(), err error) {
	noop := func() {}
	if s == nil || !s.enabled {
		return noop, nil
	}
	limits, err := s.limitsOrNil(ctx, tenantID)
	if err != nil {
		return noop, err
	}
	if limits == nil {
		return noop, nil
	}
	key := s.concurrentKey(tenantID)
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		if e := s.onRedisErr("AcquireConcurrent", err); e != nil {
			return noop, e
		}
		return noop, nil
	}
	_ = s.rdb.Expire(ctx, key, s.concurrentTTL).Err()
	if int(n) > limits.MaxConcurrentCalls {
		_, _ = s.rdb.Decr(ctx, key).Result()
		return noop, limitExceeded(DimMaxConcurrentCalls, limits.MaxConcurrentCalls, int(n)-1)
	}
	released := false
	return func() {
		if released || s.rdb == nil {
			return
		}
		released = true
		rctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		v, err := s.rdb.Decr(rctx, key).Result()
		if err == nil && v < 0 {
			_ = s.rdb.Set(rctx, key, 0, s.concurrentTTL).Err()
		}
	}, nil
}

// AddCallMinutes increments the monthly call-minutes counter.
func (s *Service) AddCallMinutes(ctx context.Context, tenantID string, minutes int) error {
	if s == nil || !s.enabled || minutes <= 0 || tenantID == "" {
		return nil
	}
	key := s.minutesKey(tenantID)
	n, err := s.rdb.IncrBy(ctx, key, int64(minutes)).Result()
	if err != nil {
		return s.onRedisErr("AddCallMinutes", err)
	}
	// Expire after ~40 days so stale months clean up.
	if n == int64(minutes) {
		_ = s.rdb.Expire(ctx, key, 40*24*time.Hour).Err()
	}
	return nil
}

// dayKey returns YYYYMMDD in the given IANA timezone (falls back to UTC).
func dayKey(now time.Time, timezone string) string {
	loc := time.UTC
	if timezone != "" {
		if l, err := time.LoadLocation(timezone); err == nil {
			loc = l
		}
	}
	return now.In(loc).Format("20060102")
}

func (s *Service) dailyKey(tenantID, timezone string) string {
	return s.prefix + "call_daily:" + tenantID + ":" + dayKey(s.now(), timezone)
}

// GetDailyCallMinutes returns minutes used today (tenant timezone day boundary).
func (s *Service) GetDailyCallMinutes(ctx context.Context, tenantID, timezone string) (int, error) {
	if s == nil || s.rdb == nil || tenantID == "" {
		return 0, nil
	}
	return s.getInt(ctx, s.dailyKey(tenantID, timezone))
}

// CheckDailyCallMinutes denies when maxPerDay > 0 and usage already at/over the cap.
// maxPerDay == 0 means unset (no operational daily cap).
func (s *Service) CheckDailyCallMinutes(ctx context.Context, tenantID, timezone string, maxPerDay int) error {
	if s == nil || !s.enabled || maxPerDay <= 0 || tenantID == "" {
		return nil
	}
	cur, err := s.GetDailyCallMinutes(ctx, tenantID, timezone)
	if err != nil {
		return s.onRedisErr("CheckDailyCallMinutes", err)
	}
	if cur >= maxPerDay {
		return DailyCallLimit(maxPerDay, cur)
	}
	return nil
}

// AddDailyCallMinutes increments the S16 daily call-minutes counter (tenant timezone day).
func (s *Service) AddDailyCallMinutes(ctx context.Context, tenantID, timezone string, minutes int) error {
	if s == nil || !s.enabled || minutes <= 0 || tenantID == "" || s.rdb == nil {
		return nil
	}
	key := s.dailyKey(tenantID, timezone)
	n, err := s.rdb.IncrBy(ctx, key, int64(minutes)).Result()
	if err != nil {
		return s.onRedisErr("AddDailyCallMinutes", err)
	}
	// ~48h TTL so day keys clean up after the boundary.
	if n == int64(minutes) {
		_ = s.rdb.Expire(ctx, key, 48*time.Hour).Err()
	}
	return nil
}

func (s *Service) previewConcurrentKey(tenantID string) string {
	return s.prefix + "preview:concurrent:" + tenantID
}

// AcquirePreviewConcurrent soft-caps concurrent tenant-admin preview voice sessions.
// Does not use package concurrent slots. Caller must release.
func (s *Service) AcquirePreviewConcurrent(ctx context.Context, tenantID string) (release func(), err error) {
	noop := func() {}
	if s == nil || tenantID == "" {
		return noop, nil
	}
	// When Redis unavailable, allow (fail-open for preview) unless rate system requires redis.
	if s.rdb == nil {
		return noop, nil
	}
	max := s.previewMaxConcurrent
	if max <= 0 {
		max = 2
	}
	key := s.previewConcurrentKey(tenantID)
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		if e := s.onRedisErr("AcquirePreviewConcurrent", err); e != nil {
			return noop, e
		}
		return noop, nil
	}
	_ = s.rdb.Expire(ctx, key, time.Hour).Err()
	if int(n) > max {
		_, _ = s.rdb.Decr(ctx, key).Result()
		return noop, PreviewConcurrent(max, int(n)-1)
	}
	released := false
	return func() {
		if released || s.rdb == nil {
			return
		}
		released = true
		rctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		v, err := s.rdb.Decr(rctx, key).Result()
		if err == nil && v < 0 {
			_ = s.rdb.Set(rctx, key, 0, time.Hour).Err()
		}
	}, nil
}

// Check is a generic dimension check (convenience for callers).
func (s *Service) Check(ctx context.Context, tenantID, dimension string) error {
	switch dimension {
	case DimMaxKMDocuments:
		return s.CheckKMDocument(ctx, tenantID)
	case DimMaxConcurrentCalls:
		// Non-mutating: read-only peek
		if s == nil || !s.enabled {
			return nil
		}
		limits, err := s.limitsOrNil(ctx, tenantID)
		if err != nil {
			return err
		}
		if limits == nil {
			return nil
		}
		cur, err := s.getInt(ctx, s.concurrentKey(tenantID))
		if err != nil {
			return s.onRedisErr("Check concurrent", err)
		}
		if cur >= limits.MaxConcurrentCalls {
			return limitExceeded(DimMaxConcurrentCalls, limits.MaxConcurrentCalls, cur)
		}
		return nil
	case DimMaxMonthlyCallMinutes:
		return s.CheckMonthlyMinutes(ctx, tenantID, 0)
	case DimMaxAIEmployees:
		n, err := s.countAvatars(ctx, tenantID)
		if err != nil {
			return err
		}
		return s.CheckAIEmployees(ctx, tenantID, n+1)
	case DimVoiceEnabled:
		return s.CheckFeature(ctx, tenantID, DimVoiceEnabled)
	case DimRAGEnabled:
		return s.CheckFeature(ctx, tenantID, DimRAGEnabled)
	default:
		return fmt.Errorf("unknown dimension %q", dimension)
	}
}

func (s *Service) rateLimitFor(bucket string) int {
	switch bucket {
	case BucketChat:
		return s.chatPerMin
	case BucketKM:
		return s.kmPerMin
	case BucketVoice:
		return s.voicePerMin
	default:
		return s.chatPerMin
	}
}

func (s *Service) concurrentKey(tenantID string) string {
	return s.prefix + "quota:" + tenantID + ":concurrent"
}

func (s *Service) minutesKey(tenantID string) string {
	ym := s.now().UTC().Format("200601")
	return s.prefix + "quota:" + tenantID + ":minutes:" + ym
}

func (s *Service) rateKey(tenantID, bucket string) string {
	win := s.now().UTC().Format("200601021504")
	return fmt.Sprintf("%srl:%s:%s:%s", s.prefix, tenantID, bucket, win)
}

func (s *Service) effective(ctx context.Context, tenantID string) (*entitlements.Effective, error) {
	if s.ents == nil {
		return nil, ErrNoEntitlement
	}
	eff, err := s.ents.GetEffective(ctx, tenantID)
	if err != nil {
		if errors.Is(err, store.ErrEntitlementNotFound) {
			return nil, ErrNoEntitlement
		}
		return nil, err
	}
	return eff, nil
}

// limitsOrNil returns limits, or nil if quota should not enforce (fail-open no entitlement).
func (s *Service) limitsOrNil(ctx context.Context, tenantID string) (*Limits, error) {
	eff, err := s.effective(ctx, tenantID)
	if err != nil {
		if errors.Is(err, ErrNoEntitlement) {
			if s.failOpen {
				return nil, nil
			}
			return nil, noEntitlement()
		}
		// entitlement store errors: fail-open if configured
		if s.failOpen {
			log.Printf("quota: entitlement error (fail-open): %v", err)
			return nil, nil
		}
		return nil, err
	}
	if eff == nil {
		if s.failOpen {
			return nil, nil
		}
		return nil, noEntitlement()
	}
	l := limitsFromRules(eff.Rules)
	return &l, nil
}

func (s *Service) collectUsage(ctx context.Context, tenantID string) (Usage, error) {
	u := Usage{}
	if n, err := s.countAvatars(ctx, tenantID); err == nil {
		u.AIEmployees = n
	} else if !s.failOpen {
		return u, err
	}
	if n, err := s.countKM(ctx, tenantID); err == nil {
		u.KMDocuments = n
	} else if !s.failOpen {
		return u, err
	}
	if s.rdb != nil {
		if n, err := s.getInt(ctx, s.concurrentKey(tenantID)); err == nil {
			u.ConcurrentCalls = n
		}
		if n, err := s.getInt(ctx, s.minutesKey(tenantID)); err == nil {
			u.MonthlyCallMinutes = n
		}
	}
	return u, nil
}

func (s *Service) countKM(ctx context.Context, tenantID string) (int, error) {
	if s.store == nil {
		return 0, nil
	}
	return s.store.CountTenantKnowledgeDocuments(ctx, tenantID)
}

func (s *Service) countAvatars(ctx context.Context, tenantID string) (int, error) {
	if s.store == nil {
		return 0, nil
	}
	return s.store.CountActiveTenantAssignments(ctx, tenantID)
}

func (s *Service) getInt(ctx context.Context, key string) (int, error) {
	if s.rdb == nil {
		return 0, nil
	}
	n, err := s.rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return n, err
}

func (s *Service) onRedisErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if s.failOpen {
		log.Printf("quota: %s redis error (fail-open): %v", op, err)
		return nil
	}
	return err
}

func limitsFromRules(rules map[string]any) Limits {
	return Limits{
		MaxAIEmployees:        intRule(rules, DimMaxAIEmployees, 0),
		MaxMonthlyCallMinutes: intRule(rules, DimMaxMonthlyCallMinutes, 0),
		MaxKMDocuments:        intRule(rules, DimMaxKMDocuments, 0),
		MaxConcurrentCalls:    intRule(rules, DimMaxConcurrentCalls, 0),
		VoiceEnabled:          boolRule(rules, DimVoiceEnabled, true),
		RAGEnabled:            boolRule(rules, DimRAGEnabled, true),
	}
}

func intRule(rules map[string]any, key string, def int) int {
	if rules == nil {
		return def
	}
	v, ok := rules[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case float32:
		return int(t)
	default:
		return def
	}
}

func boolRule(rules map[string]any, key string, def bool) bool {
	if rules == nil {
		return def
	}
	v, ok := rules[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case bool:
		return t
	default:
		return def
	}
}
