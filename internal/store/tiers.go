package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrTierNotFound  = errors.New("tier not found")
	ErrGroupNotFound = errors.New("group not found")
	ErrInvalidSlug   = errors.New("invalid slug")
	ErrSlugTaken     = errors.New("slug already exists")
)

var slugRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// CustomerTier is one row of callcenter.customer_tiers.
type CustomerTier struct {
	ID                   string    `json:"id"`
	TenantID             string    `json:"tenant_id"`
	Name                 string    `json:"name"`
	Slug                 string    `json:"slug"`
	Priority             int       `json:"priority"`
	Description          string    `json:"description,omitempty"`
	DefaultAgentID       string    `json:"default_agent_id,omitempty"`
	AIReplyLocale        string    `json:"ai_reply_locale,omitempty"`
	MaxMinutesPerCall    int       `json:"max_minutes_per_call"`
	MaxCallMinutesPerDay int       `json:"max_call_minutes_per_day"`
	Active               bool      `json:"active"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
	UpdatedAt            time.Time `json:"updated_at,omitempty"`
}

// CustomerGroup is one row of callcenter.customer_groups.
type CustomerGroup struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func (s *Store) ensureTiersSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_tiers (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  name text NOT NULL,
  slug text NOT NULL,
  priority integer NOT NULL DEFAULT 0,
  description text NOT NULL DEFAULT '',
  default_agent_id text,
  ai_reply_locale text NOT NULL DEFAULT '',
  max_minutes_per_call integer NOT NULL DEFAULT 0,
  max_call_minutes_per_day integer NOT NULL DEFAULT 0,
  active boolean NOT NULL DEFAULT true,%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS customer_tiers_tenant_slug_uidx
ON %s.customer_tiers (tenant_id, slug)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS customer_tiers_tenant_idx
ON %s.customer_tiers (tenant_id)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_groups (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  name text NOT NULL,
  slug text NOT NULL,
  description text NOT NULL DEFAULT '',%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS customer_groups_tenant_slug_uidx
ON %s.customer_groups (tenant_id, slug)`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("tiers schema: %w", err)
		}
	}
	return nil
}

// NormalizeSlug lowercases and validates slug form.
func NormalizeSlug(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, "_", "-")
	v = strings.ReplaceAll(v, " ", "-")
	for strings.Contains(v, "--") {
		v = strings.ReplaceAll(v, "--", "-")
	}
	v = strings.Trim(v, "-")
	if v == "" || !slugRE.MatchString(v) || len(v) > 64 {
		return "", ErrInvalidSlug
	}
	return v, nil
}

func newTierID() (string, error) {
	var b [10]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "tier_" + hex.EncodeToString(b[:]), nil
}

func newGroupID() (string, error) {
	var b [10]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "grp_" + hex.EncodeToString(b[:]), nil
}

func (s *Store) scanTier(row pgx.Row) (*CustomerTier, error) {
	var t CustomerTier
	var defAgent *string
	err := row.Scan(
		&t.ID, &t.TenantID, &t.Name, &t.Slug, &t.Priority, &t.Description,
		&defAgent, &t.AIReplyLocale, &t.MaxMinutesPerCall, &t.MaxCallMinutesPerDay,
		&t.Active, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if defAgent != nil {
		t.DefaultAgentID = *defAgent
	}
	return &t, nil
}

func (s *Store) ListCustomerTiers(ctx context.Context, tenantID string) ([]CustomerTier, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT id, tenant_id, name, slug, priority, COALESCE(description,''), default_agent_id,
       COALESCE(ai_reply_locale,''), max_minutes_per_call, max_call_minutes_per_day,
       active, created_at, updated_at
FROM %s.customer_tiers WHERE tenant_id = $1
ORDER BY priority DESC, name ASC`, schema), tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CustomerTier
	for rows.Next() {
		var t CustomerTier
		var defAgent *string
		if err := rows.Scan(
			&t.ID, &t.TenantID, &t.Name, &t.Slug, &t.Priority, &t.Description,
			&defAgent, &t.AIReplyLocale, &t.MaxMinutesPerCall, &t.MaxCallMinutesPerDay,
			&t.Active, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if defAgent != nil {
			t.DefaultAgentID = *defAgent
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) GetCustomerTier(ctx context.Context, tenantID, id string) (*CustomerTier, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	t, err := s.scanTier(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, tenant_id, name, slug, priority, COALESCE(description,''), default_agent_id,
       COALESCE(ai_reply_locale,''), max_minutes_per_call, max_call_minutes_per_day,
       active, created_at, updated_at
FROM %s.customer_tiers WHERE tenant_id = $1 AND id = $2`, schema), tenantID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTierNotFound
	}
	return t, err
}

// CreateCustomerTierInput for insert.
type CreateCustomerTierInput struct {
	Name                 string
	Slug                 string
	Priority             int
	Description          string
	DefaultAgentID       string
	AIReplyLocale        string
	MaxMinutesPerCall    int
	MaxCallMinutesPerDay int
	Active               *bool
}

func (s *Store) CreateCustomerTier(ctx context.Context, tenantID string, in CreateCustomerTierInput) (*CustomerTier, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	slug, err := NormalizeSlug(in.Slug)
	if err != nil {
		// derive from name
		slug, err = NormalizeSlug(name)
		if err != nil {
			return nil, ErrInvalidSlug
		}
	}
	locale, err := NormalizeOptionalLocale(in.AIReplyLocale)
	if err != nil {
		return nil, ErrInvalidLocale
	}
	if in.MaxMinutesPerCall < 0 || in.MaxCallMinutesPerDay < 0 {
		return nil, fmt.Errorf("call limits must be >= 0")
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	id, err := newTierID()
	if err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	var agent any
	if strings.TrimSpace(in.DefaultAgentID) == "" {
		agent = nil
	} else {
		agent = strings.TrimSpace(in.DefaultAgentID)
	}
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.customer_tiers (
  id, tenant_id, name, slug, priority, description, default_agent_id, ai_reply_locale,
  max_minutes_per_call, max_call_minutes_per_day, active, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12)`, schema),
		id, tenantID, name, slug, in.Priority, strings.TrimSpace(in.Description), agent, locale,
		in.MaxMinutesPerCall, in.MaxCallMinutesPerDay, active, actor)
	if err != nil {
		if strings.Contains(err.Error(), "customer_tiers_tenant_slug") || strings.Contains(err.Error(), "unique") {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	return s.GetCustomerTier(ctx, tenantID, id)
}

// UpdateCustomerTierInput partial update.
type UpdateCustomerTierInput struct {
	Name                 *string
	Slug                 *string
	Priority             *int
	Description          *string
	DefaultAgentID       *string
	AIReplyLocale        *string
	MaxMinutesPerCall    *int
	MaxCallMinutesPerDay *int
	Active               *bool
}

func (s *Store) UpdateCustomerTier(ctx context.Context, tenantID, id string, in UpdateCustomerTierInput) (*CustomerTier, error) {
	cur, err := s.GetCustomerTier(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	name := cur.Name
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, fmt.Errorf("name is required")
		}
	}
	slug := cur.Slug
	if in.Slug != nil {
		slug, err = NormalizeSlug(*in.Slug)
		if err != nil {
			return nil, err
		}
	}
	priority := cur.Priority
	if in.Priority != nil {
		priority = *in.Priority
	}
	desc := cur.Description
	if in.Description != nil {
		desc = strings.TrimSpace(*in.Description)
	}
	agentID := cur.DefaultAgentID
	if in.DefaultAgentID != nil {
		agentID = strings.TrimSpace(*in.DefaultAgentID)
	}
	locale := cur.AIReplyLocale
	if in.AIReplyLocale != nil {
		locale, err = NormalizeOptionalLocale(*in.AIReplyLocale)
		if err != nil {
			return nil, err
		}
	}
	maxCall := cur.MaxMinutesPerCall
	if in.MaxMinutesPerCall != nil {
		if *in.MaxMinutesPerCall < 0 {
			return nil, fmt.Errorf("call limits must be >= 0")
		}
		maxCall = *in.MaxMinutesPerCall
	}
	maxDay := cur.MaxCallMinutesPerDay
	if in.MaxCallMinutesPerDay != nil {
		if *in.MaxCallMinutesPerDay < 0 {
			return nil, fmt.Errorf("call limits must be >= 0")
		}
		maxDay = *in.MaxCallMinutesPerDay
	}
	active := cur.Active
	if in.Active != nil {
		active = *in.Active
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	var agent any
	if agentID == "" {
		agent = nil
	} else {
		agent = agentID
	}
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.customer_tiers SET
  name=$3, slug=$4, priority=$5, description=$6, default_agent_id=$7, ai_reply_locale=$8,
  max_minutes_per_call=$9, max_call_minutes_per_day=$10, active=$11,
  updated_by=$12, updated_at=now()
WHERE tenant_id=$1 AND id=$2`, schema),
		tenantID, id, name, slug, priority, desc, agent, locale, maxCall, maxDay, active, actor)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrTierNotFound
	}
	return s.GetCustomerTier(ctx, tenantID, id)
}

func (s *Store) DeleteCustomerTier(ctx context.Context, tenantID, id string) error {
	if s.pg == nil {
		return fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s.customer_tiers WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTierNotFound
	}
	return nil
}

// --- groups ---

func (s *Store) ListCustomerGroups(ctx context.Context, tenantID string) ([]CustomerGroup, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT id, tenant_id, name, slug, COALESCE(description,''), created_at, updated_at
FROM %s.customer_groups WHERE tenant_id=$1 ORDER BY name ASC`, schema), tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CustomerGroup
	for rows.Next() {
		var g CustomerGroup
		if err := rows.Scan(&g.ID, &g.TenantID, &g.Name, &g.Slug, &g.Description, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) GetCustomerGroup(ctx context.Context, tenantID, id string) (*CustomerGroup, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var g CustomerGroup
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, tenant_id, name, slug, COALESCE(description,''), created_at, updated_at
FROM %s.customer_groups WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id).Scan(
		&g.ID, &g.TenantID, &g.Name, &g.Slug, &g.Description, &g.CreatedAt, &g.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrGroupNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *Store) CreateCustomerGroup(ctx context.Context, tenantID, name, slug, description string) (*CustomerGroup, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	sl, err := NormalizeSlug(slug)
	if err != nil {
		sl, err = NormalizeSlug(name)
		if err != nil {
			return nil, ErrInvalidSlug
		}
	}
	id, err := newGroupID()
	if err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.customer_groups (id, tenant_id, name, slug, description, created_by, updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$6)`, schema), id, tenantID, name, sl, strings.TrimSpace(description), actor)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	return s.GetCustomerGroup(ctx, tenantID, id)
}

func (s *Store) UpdateCustomerGroup(ctx context.Context, tenantID, id, name, slug, description string) (*CustomerGroup, error) {
	if _, err := s.GetCustomerGroup(ctx, tenantID, id); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	sl, err := NormalizeSlug(slug)
	if err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.customer_groups SET name=$3, slug=$4, description=$5, updated_by=$6, updated_at=now()
WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id, name, sl, strings.TrimSpace(description), actor)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrGroupNotFound
	}
	return s.GetCustomerGroup(ctx, tenantID, id)
}

func (s *Store) DeleteCustomerGroup(ctx context.Context, tenantID, id string) error {
	if s.pg == nil {
		return fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s.customer_groups WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrGroupNotFound
	}
	return nil
}
