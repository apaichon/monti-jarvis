package entitlements

import (
	"context"
	"encoding/json"
	"time"

	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/redis/go-redis/v9"
)

type PackageSummary struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Effective struct {
	TenantID      string         `json:"tenant_id"`
	Package       PackageSummary `json:"package"`
	Status        string         `json:"status"`
	RulesSchemaID string         `json:"rules_schema_id"`
	Rules         map[string]any `json:"rules"`
	ValidFrom     time.Time      `json:"valid_from"`
	ValidUntil    *time.Time     `json:"valid_until"`
}

type Service struct {
	store   *store.Store
	rdb     *redis.Client
	prefix  string
	ttl     time.Duration
	enabled bool
}

func New(st *store.Store, cfg env.Config) *Service {
	return &Service{
		store:   st,
		rdb:     st.Redis(),
		prefix:  cfg.RedisPrefix,
		ttl:     cfg.EntitlementCacheTTL,
		enabled: cfg.EntitlementCacheEnabled && st != nil && st.Redis() != nil,
	}
}

func (s *Service) CacheStatus() string {
	if s == nil || !s.enabled {
		return "disabled"
	}
	return "ok"
}

func (s *Service) cacheKey(tenantID string) string {
	return s.prefix + "entitlement:" + tenantID
}

func (s *Service) GetEffective(ctx context.Context, tenantID string) (*Effective, error) {
	if s.enabled {
		raw, err := s.rdb.Get(ctx, s.cacheKey(tenantID)).Result()
		if err == nil {
			var eff Effective
			if json.Unmarshal([]byte(raw), &eff) == nil {
				return &eff, nil
			}
		}
	}
	ent, err := s.store.GetActiveEntitlement(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	eff := fromStore(ent)
	if s.enabled {
		if b, err := json.Marshal(eff); err == nil {
			_ = s.rdb.Set(ctx, s.cacheKey(tenantID), b, s.ttl).Err()
		}
	}
	return eff, nil
}

func (s *Service) Invalidate(ctx context.Context, tenantID string) {
	if s.enabled {
		_ = s.rdb.Del(ctx, s.cacheKey(tenantID)).Err()
	}
}

func fromStore(ent *store.TenantEntitlement) *Effective {
	pkg := PackageSummary{}
	if ent.Package != nil {
		pkg = PackageSummary{ID: ent.Package.ID, Slug: ent.Package.Slug, Name: ent.Package.Name}
	}
	return &Effective{
		TenantID:      ent.TenantID,
		Package:       pkg,
		Status:        ent.Status,
		RulesSchemaID: ent.RulesSchemaID,
		Rules:         ent.RulesSnapshot,
		ValidFrom:     ent.ValidFrom,
		ValidUntil:    ent.ValidUntil,
	}
}