package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrPackageNotFound      = errors.New("package not found")
	ErrRuleSchemaNotFound   = errors.New("rules schema not found")
	ErrEntitlementNotFound  = errors.New("entitlement not found")
	ErrTenantNotFound       = errors.New("tenant not found")
	ErrActiveEntitlement    = errors.New("active entitlement exists")
	ErrPackageHasEntitlements = errors.New("package has active entitlements")
)

type RuleSchema struct {
	ID      string
	Version int
	Name    string
	Fields  json.RawMessage
	Status  string
}

type Package struct {
	ID             string
	Slug           string
	Name           string
	Description    string
	Status         string
	PriceCents     int
	Currency       string
	BillingPeriod  string
	RulesSchemaID  string
	Rules          map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TenantEntitlement struct {
	ID             string
	TenantID       string
	PackageID      string
	RulesSchemaID  string
	RulesSnapshot  map[string]any
	Status         string
	ValidFrom      time.Time
	ValidUntil     *time.Time
	Package        *Package
}

const rulesV1Fields = `{
  "max_ai_employees": {"type":"int","min":0,"required":true,"description":"Max AI avatars"},
  "max_monthly_call_minutes": {"type":"int","min":0,"required":true,"description":"Monthly voice minutes"},
  "max_km_documents": {"type":"int","min":0,"required":true,"description":"KM documents"},
  "max_concurrent_calls": {"type":"int","min":0,"required":true,"description":"Parallel calls"},
  "voice_enabled": {"type":"bool","required":true,"default":true},
  "rag_enabled": {"type":"bool","required":true,"default":true}
}`

func (s *Store) ensurePackagesSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.package_rule_schemas (
  id text PRIMARY KEY,
  version int NOT NULL UNIQUE,
  name text NOT NULL,
  fields jsonb NOT NULL,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deprecated')),%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.packages (
  id text PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
  price_cents int NOT NULL DEFAULT 0,
  currency text NOT NULL DEFAULT 'USD',
  billing_period text NOT NULL DEFAULT 'monthly',%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.package_limits (
  package_id text PRIMARY KEY REFERENCES %s.packages(id) ON DELETE CASCADE,
  rules_schema_id text NOT NULL REFERENCES %s.package_rule_schemas(id),
  rules jsonb NOT NULL,%s
)`, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_entitlements (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  package_id text NOT NULL REFERENCES %s.packages(id),
  rules_schema_id text NOT NULL REFERENCES %s.package_rule_schemas(id),
  rules_snapshot jsonb NOT NULL,
  status text NOT NULL CHECK (status IN ('active', 'suspended', 'revoked', 'expired')),
  valid_from timestamptz NOT NULL DEFAULT now(),
  valid_until timestamptz,%s
)`, schema, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS tenant_entitlements_one_active_idx
ON %s.tenant_entitlements (tenant_id) WHERE status = 'active'`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.embedding_models (
  id text PRIMARY KEY,
  provider text NOT NULL,
  model_key text NOT NULL,
  dimensions int NOT NULL DEFAULT 768,
  status text NOT NULL DEFAULT 'active',%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.voice_providers (
  id text PRIMARY KEY,
  provider text NOT NULL,
  model_key text NOT NULL,
  modality text NOT NULL DEFAULT 'audio',
  status text NOT NULL DEFAULT 'active',%s
)`, schema, auditColumnsDDL),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return s.seedPackages(ctx)
}

func (s *Store) seedPackages(ctx context.Context) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	demoTenant := s.cfg.DemoTenantID
	if demoTenant == "" {
		demoTenant = "demo"
	}

	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.package_rule_schemas (id, version, name, fields, status)
VALUES ('rules-v1', 1, 'Sprint 4 base limits', $1::jsonb, 'active')
ON CONFLICT (id) DO NOTHING`, schema), rulesV1Fields)
	if err != nil {
		return err
	}

	pkgs := []struct {
		id, slug, name string
		rules          string
	}{
		{"pkg-starter", "starter", "Starter", `{"max_ai_employees":2,"max_monthly_call_minutes":500,"max_km_documents":50,"max_concurrent_calls":2,"voice_enabled":true,"rag_enabled":true}`},
		{"pkg-pro", "pro", "Pro", `{"max_ai_employees":10,"max_monthly_call_minutes":5000,"max_km_documents":500,"max_concurrent_calls":10,"voice_enabled":true,"rag_enabled":true}`},
		{"pkg-enterprise", "enterprise", "Enterprise", `{"max_ai_employees":50,"max_monthly_call_minutes":50000,"max_km_documents":5000,"max_concurrent_calls":50,"voice_enabled":true,"rag_enabled":true}`},
	}
	for _, p := range pkgs {
		_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.packages (id, slug, name, status, price_cents, currency, billing_period)
VALUES ($1, $2, $3, 'active', 0, 'USD', 'monthly')
ON CONFLICT (id) DO NOTHING`, schema), p.id, p.slug, p.name)
		if err != nil {
			return err
		}
		_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.package_limits (package_id, rules_schema_id, rules)
VALUES ($1, 'rules-v1', $2::jsonb)
ON CONFLICT (package_id) DO NOTHING`, schema), p.id, p.rules)
		if err != nil {
			return err
		}
	}

	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_entitlements (id, tenant_id, package_id, rules_schema_id, rules_snapshot, status)
SELECT 'ent_demo_starter', $1, 'pkg-starter', 'rules-v1', pl.rules, 'active'
FROM %s.package_limits pl WHERE pl.package_id = 'pkg-starter'
ON CONFLICT (id) DO NOTHING`, schema, schema), demoTenant)
	if err != nil {
		return err
	}

	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.embedding_models (id, provider, model_key, dimensions, status)
VALUES ('emb-gemini-001', 'google', 'gemini-embedding-001', 768, 'active')
ON CONFLICT (id) DO NOTHING`, schema))
	if err != nil {
		return err
	}

	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.voice_providers (id, provider, model_key, modality, status)
VALUES ('voice-gemini-live', 'google', 'gemini-2.5-flash-native-audio-latest', 'audio', 'active')
ON CONFLICT (id) DO NOTHING`, schema))
	return err
}

func (s *Store) ListRuleSchemas(ctx context.Context, status string) ([]RuleSchema, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	q := fmt.Sprintf(`SELECT id, version, name, fields, status FROM %s.package_rule_schemas`, schema)
	args := []any{}
	if status != "" {
		q += ` WHERE status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY version`
	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RuleSchema
	for rows.Next() {
		var rs RuleSchema
		if err := rows.Scan(&rs.ID, &rs.Version, &rs.Name, &rs.Fields, &rs.Status); err != nil {
			return nil, err
		}
		out = append(out, rs)
	}
	return out, rows.Err()
}

func (s *Store) GetRuleSchema(ctx context.Context, id string) (*RuleSchema, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var rs RuleSchema
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, version, name, fields, status FROM %s.package_rule_schemas WHERE id = $1`, schema), id).
		Scan(&rs.ID, &rs.Version, &rs.Name, &rs.Fields, &rs.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleSchemaNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

func (s *Store) ListPackages(ctx context.Context, status string) ([]Package, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	q := fmt.Sprintf(`
SELECT p.id, p.slug, p.name, p.description, p.status, p.price_cents, p.currency, p.billing_period,
       pl.rules_schema_id, pl.rules, p.created_at, p.updated_at
FROM %s.packages p
JOIN %s.package_limits pl ON pl.package_id = p.id`, schema, schema)
	args := []any{}
	if status != "" {
		q += ` WHERE p.status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY p.slug`
	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPackages(rows)
}

func scanPackages(rows pgx.Rows) ([]Package, error) {
	var out []Package
	for rows.Next() {
		var p Package
		var rulesRaw []byte
		if err := rows.Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.Status, &p.PriceCents, &p.Currency, &p.BillingPeriod,
			&p.RulesSchemaID, &rulesRaw, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rulesRaw, &p.Rules)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetPackage(ctx context.Context, id string) (*Package, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var p Package
	var rulesRaw []byte
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT p.id, p.slug, p.name, p.description, p.status, p.price_cents, p.currency, p.billing_period,
       pl.rules_schema_id, pl.rules, p.created_at, p.updated_at
FROM %s.packages p
JOIN %s.package_limits pl ON pl.package_id = p.id
WHERE p.id = $1`, schema, schema), id).
		Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.Status, &p.PriceCents, &p.Currency, &p.BillingPeriod,
			&p.RulesSchemaID, &rulesRaw, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPackageNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(rulesRaw, &p.Rules)
	return &p, nil
}

func (s *Store) CreatePackage(ctx context.Context, p Package) (*Package, error) {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	rulesJSON, err := json.Marshal(p.Rules)
	if err != nil {
		return nil, err
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.packages (id, slug, name, description, status, price_cents, currency, billing_period, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)`, schema),
		p.ID, p.Slug, p.Name, p.Description, p.Status, p.PriceCents, p.Currency, p.BillingPeriod, actor)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.package_limits (package_id, rules_schema_id, rules, created_by, updated_by)
VALUES ($1, $2, $3::jsonb, $4, $4)`, schema),
		p.ID, p.RulesSchemaID, string(rulesJSON), actor)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetPackage(ctx, p.ID)
}

func (s *Store) UpdatePackage(ctx context.Context, p Package) (*Package, error) {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	rulesJSON, err := json.Marshal(p.Rules)
	if err != nil {
		return nil, err
	}
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.packages
SET slug = $2, name = $3, description = $4, status = $5, price_cents = $6, currency = $7, billing_period = $8, updated_by = $9
WHERE id = $1`, schema),
		p.ID, p.Slug, p.Name, p.Description, p.Status, p.PriceCents, p.Currency, p.BillingPeriod, actor)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrPackageNotFound
	}
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.package_limits SET rules_schema_id = $2, rules = $3::jsonb, updated_by = $4 WHERE package_id = $1`, schema),
		p.ID, p.RulesSchemaID, string(rulesJSON), actor)
	if err != nil {
		return nil, err
	}
	return s.GetPackage(ctx, p.ID)
}

func (s *Store) ArchivePackage(ctx context.Context, id string) error {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	var active int
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s.tenant_entitlements WHERE package_id = $1 AND status = 'active'`, schema), id).Scan(&active)
	if err != nil {
		return err
	}
	if active > 0 {
		return ErrPackageHasEntitlements
	}
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.packages SET status = 'archived', updated_by = $2 WHERE id = $1`, schema), id, actor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPackageNotFound
	}
	return nil
}

func (s *Store) CountActiveEntitlementsForPackage(ctx context.Context, packageID string) (int, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var n int
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s.tenant_entitlements WHERE package_id = $1 AND status = 'active'`, schema), packageID).Scan(&n)
	return n, err
}

func (s *Store) TenantExists(ctx context.Context, tenantID string) (bool, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var n int
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.tenants WHERE id = $1`, schema), tenantID).Scan(&n)
	return n > 0, err
}

func (s *Store) GetActiveEntitlement(ctx context.Context, tenantID string) (*TenantEntitlement, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	return s.scanEntitlement(ctx, fmt.Sprintf(`
SELECT e.id, e.tenant_id, e.package_id, e.rules_schema_id, e.rules_snapshot, e.status, e.valid_from, e.valid_until,
       p.id, p.slug, p.name, p.description, p.status, p.price_cents, p.currency, p.billing_period,
       pl.rules_schema_id, pl.rules, p.created_at, p.updated_at
FROM %s.tenant_entitlements e
JOIN %s.packages p ON p.id = e.package_id
JOIN %s.package_limits pl ON pl.package_id = p.id
WHERE e.tenant_id = $1 AND e.status = 'active'
ORDER BY e.valid_from DESC LIMIT 1`, schema, schema, schema), tenantID)
}

func scanEntitlementRow(row pgx.Row) (*TenantEntitlement, error) {
	var e TenantEntitlement
	var snapRaw, rulesRaw []byte
	var pkg Package
	var validUntil *time.Time
	err := row.Scan(
		&e.ID, &e.TenantID, &e.PackageID, &e.RulesSchemaID, &snapRaw, &e.Status, &e.ValidFrom, &validUntil,
		&pkg.ID, &pkg.Slug, &pkg.Name, &pkg.Description, &pkg.Status, &pkg.PriceCents, &pkg.Currency, &pkg.BillingPeriod,
		&pkg.RulesSchemaID, &rulesRaw, &pkg.CreatedAt, &pkg.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEntitlementNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(snapRaw, &e.RulesSnapshot)
	_ = json.Unmarshal(rulesRaw, &pkg.Rules)
	e.ValidUntil = validUntil
	e.Package = &pkg
	return &e, nil
}

func (s *Store) scanEntitlement(ctx context.Context, q string, tenantID string) (*TenantEntitlement, error) {
	return scanEntitlementRow(s.pg.QueryRow(ctx, q, tenantID))
}

func (s *Store) AssignEntitlement(ctx context.Context, tenantID, packageID string) (*TenantEntitlement, error) {
	actor := auditctx.ActorID(ctx)
	exists, err := s.TenantExists(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTenantNotFound
	}
	pkg, err := s.GetPackage(ctx, packageID)
	if err != nil {
		return nil, err
	}
	if pkg.Status == "archived" {
		return nil, ErrPackageNotFound
	}
	snapJSON, err := json.Marshal(pkg.Rules)
	if err != nil {
		return nil, err
	}
	entID := "ent_" + tenantID + "_" + packageID
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_entitlements SET status = 'revoked', updated_by = $2
WHERE tenant_id = $1 AND status = 'active'`, schema), tenantID, actor)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_entitlements (id, tenant_id, package_id, rules_schema_id, rules_snapshot, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5::jsonb, 'active', $6, $6)`, schema),
		entID, tenantID, packageID, pkg.RulesSchemaID, string(snapJSON), actor)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetActiveEntitlement(ctx, tenantID)
}

func (s *Store) RevokeEntitlement(ctx context.Context, tenantID string) error {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_entitlements SET status = 'revoked', updated_by = $2
WHERE tenant_id = $1 AND status = 'active'`, schema), tenantID, actor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrEntitlementNotFound
	}
	return nil
}