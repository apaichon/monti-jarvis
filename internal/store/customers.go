package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrCustomerConflict   = errors.New("customer identity conflict")
	ErrDomainRuleNotFound = errors.New("domain rule not found")
	ErrDomainRuleTaken    = errors.New("domain rule already exists")
	ErrImportNotFound     = errors.New("customer import not found")
)

var sourceRE = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
var domainLabelRE = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

type Customer struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"-"`
	Email       string         `json:"email,omitempty"`
	Phone       string         `json:"phone,omitempty"`
	DisplayName string         `json:"display_name"`
	Locale      string         `json:"locale,omitempty"`
	TierID      string         `json:"tier_id,omitempty"`
	GroupIDs    []string       `json:"group_ids"`
	Source      string         `json:"source"`
	ExternalID  string         `json:"external_id,omitempty"`
	Status      string         `json:"status"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type CustomerInput struct {
	Email       string
	Phone       string
	DisplayName string
	Locale      string
	TierID      string
	GroupIDs    []string
	Source      string
	ExternalID  string
	Metadata    map[string]any
	Status      string
}

type CustomerListFilter struct {
	Query  string
	Status string
	TierID string
	Limit  int
}

type CustomerUpsertResult struct {
	Customer *Customer
	Outcome  string
}

type CustomerImportJob struct {
	ID           string           `json:"id"`
	TenantID     string           `json:"-"`
	Filename     string           `json:"filename"`
	Mode         string           `json:"mode"`
	Status       string           `json:"status"`
	TotalRows    int              `json:"total_rows"`
	AcceptedRows int              `json:"accepted_rows"`
	CreatedRows  int              `json:"created_rows"`
	UpdatedRows  int              `json:"updated_rows"`
	RejectedRows int              `json:"rejected_rows"`
	Errors       []map[string]any `json:"errors"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type CustomerDomainRule struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"-"`
	Domain         string    `json:"domain"`
	Policy         string    `json:"policy"`
	DefaultTierID  string    `json:"default_tier_id,omitempty"`
	DefaultGroupID string    `json:"default_group_id,omitempty"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CustomerDomainRuleInput struct {
	Domain         string
	Policy         string
	DefaultTierID  string
	DefaultGroupID string
	Active         *bool
}

func (s *Store) ensureCustomersSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customers (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  email text NOT NULL DEFAULT '',
  email_normalized text,
  phone text NOT NULL DEFAULT '',
  display_name text NOT NULL,
  locale text NOT NULL DEFAULT '',
  tier_id text REFERENCES %s.customer_tiers(id) ON DELETE RESTRICT,
  source text NOT NULL DEFAULT 'manual',
  external_id text,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,%s
)`, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS customers_tenant_email_uidx
ON %s.customers (tenant_id, email_normalized) WHERE email_normalized IS NOT NULL`, schema),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS customers_tenant_external_uidx
ON %s.customers (tenant_id, source, external_id) WHERE external_id IS NOT NULL AND external_id <> ''`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS customers_tenant_status_idx
ON %s.customers (tenant_id, status, updated_at DESC)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_group_members (
  customer_id text NOT NULL REFERENCES %s.customers(id) ON DELETE CASCADE,
  group_id text NOT NULL REFERENCES %s.customer_groups(id) ON DELETE RESTRICT,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,%s,
  PRIMARY KEY (customer_id, group_id)
)`, schema, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS customer_group_members_tenant_idx
ON %s.customer_group_members (tenant_id, group_id)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_import_jobs (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  filename text NOT NULL DEFAULT '',
  mode text NOT NULL CHECK (mode IN ('dry_run','commit')),
  status text NOT NULL CHECK (status IN ('validating','validated','completed','failed')),
  total_rows integer NOT NULL DEFAULT 0,
  created_rows integer NOT NULL DEFAULT 0,
  updated_rows integer NOT NULL DEFAULT 0,
  rejected_rows integer NOT NULL DEFAULT 0,
  errors jsonb NOT NULL DEFAULT '[]'::jsonb,%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS customer_import_jobs_tenant_idx
ON %s.customer_import_jobs (tenant_id, created_at DESC)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_domain_rules (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  domain text NOT NULL,
  policy text NOT NULL CHECK (policy IN ('allow','deny')),
  default_tier_id text REFERENCES %s.customer_tiers(id) ON DELETE RESTRICT,
  default_group_id text REFERENCES %s.customer_groups(id) ON DELETE RESTRICT,
  active boolean NOT NULL DEFAULT true,%s
)`, schema, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS customer_domain_rules_tenant_domain_uidx
ON %s.customer_domain_rules (tenant_id, domain)`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("customers schema: %w", err)
		}
	}
	return nil
}

func NormalizeCustomerEmail(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	address, err := mail.ParseAddress(value)
	if err != nil || !strings.Contains(address.Address, "@") {
		return "", fmt.Errorf("invalid email")
	}
	return strings.ToLower(address.Address), nil
}

func NormalizeCustomerSource(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		value = "manual"
	}
	if !sourceRE.MatchString(value) {
		return "", fmt.Errorf("invalid source")
	}
	return value, nil
}

func NormalizeCustomerDomain(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimSuffix(value, ".")
	if value == "" || strings.ContainsAny(value, "/:@ ") || !strings.Contains(value, ".") || len(value) > 253 {
		return "", fmt.Errorf("invalid domain")
	}
	for _, label := range strings.Split(value, ".") {
		if !domainLabelRE.MatchString(label) {
			return "", fmt.Errorf("invalid domain")
		}
	}
	return value, nil
}

func (s *Store) validateCustomerAssignments(ctx context.Context, tenantID, tierID string, groupIDs []string) error {
	if tierID != "" {
		if _, err := s.GetCustomerTier(ctx, tenantID, tierID); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for _, id := range groupIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		if _, err := s.GetCustomerGroup(ctx, tenantID, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) prepareCustomerInput(ctx context.Context, tenantID string, in CustomerInput) (CustomerInput, string, error) {
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if in.DisplayName == "" || len(in.DisplayName) > 200 {
		return in, "", fmt.Errorf("display_name is required and must be at most 200 characters")
	}
	emailNormalized, err := NormalizeCustomerEmail(in.Email)
	if err != nil {
		return in, "", err
	}
	in.Email = emailNormalized
	in.Phone = strings.TrimSpace(in.Phone)
	if len(in.Phone) > 40 {
		return in, "", fmt.Errorf("phone must be at most 40 characters")
	}
	in.Locale, err = NormalizeOptionalLocale(in.Locale)
	if err != nil {
		return in, "", err
	}
	in.Source, err = NormalizeCustomerSource(in.Source)
	if err != nil {
		return in, "", err
	}
	in.ExternalID = strings.TrimSpace(in.ExternalID)
	if emailNormalized == "" && in.ExternalID == "" {
		return in, "", fmt.Errorf("email or external_id is required")
	}
	if in.Status == "" {
		in.Status = "active"
	}
	if in.Status != "active" && in.Status != "inactive" {
		return in, "", fmt.Errorf("status must be active or inactive")
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	if containsCredentialMetadata(in.Metadata) {
		return in, "", fmt.Errorf("metadata must not contain credential or token fields")
	}
	metadata, err := json.Marshal(in.Metadata)
	if err != nil || len(metadata) > 16*1024 {
		return in, "", fmt.Errorf("metadata must be valid JSON under 16 KiB")
	}
	if in.TierID == "" && len(in.GroupIDs) == 0 && emailNormalized != "" {
		if at := strings.LastIndex(emailNormalized, "@"); at >= 0 {
			if rule, findErr := s.FindCustomerDomainRule(ctx, tenantID, emailNormalized[at+1:]); findErr == nil && rule.Active {
				in.TierID = rule.DefaultTierID
				if rule.DefaultGroupID != "" {
					in.GroupIDs = []string{rule.DefaultGroupID}
				}
			}
		}
	}
	if err := s.validateCustomerAssignments(ctx, tenantID, in.TierID, in.GroupIDs); err != nil {
		return in, "", err
	}
	return in, emailNormalized, nil
}

func containsCredentialMetadata(value map[string]any) bool {
	blocked := map[string]bool{
		"password": true, "password_hash": true, "otp": true, "secret": true,
		"token": true, "access_token": true, "refresh_token": true,
	}
	for key, item := range value {
		if blocked[strings.ToLower(strings.TrimSpace(key))] {
			return true
		}
		if nested, ok := item.(map[string]any); ok && containsCredentialMetadata(nested) {
			return true
		}
	}
	return false
}

// ValidateCustomerInput applies the same normalization/reference checks as a write without mutation.
func (s *Store) ValidateCustomerInput(ctx context.Context, tenantID string, in CustomerInput) (CustomerInput, error) {
	prepared, _, err := s.prepareCustomerInput(ctx, tenantID, in)
	return prepared, err
}

func (s *Store) UpsertCustomer(ctx context.Context, tenantID string, in CustomerInput) (*CustomerUpsertResult, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	in, emailNormalized, err := s.prepareCustomerInput(ctx, tenantID, in)
	if err != nil {
		return nil, err
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	id, outcome, err := s.writeCustomerTx(ctx, tx, tenantID, "", in, emailNormalized)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	row, err := s.GetCustomer(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	return &CustomerUpsertResult{Customer: row, Outcome: outcome}, nil
}

// UpdateCustomer updates the requested tenant-owned id directly and never redirects the write
// to another identity match.
func (s *Store) UpdateCustomer(ctx context.Context, tenantID, id string, in CustomerInput) (*Customer, error) {
	if _, err := s.GetCustomer(ctx, tenantID, id); err != nil {
		return nil, err
	}
	in, emailNormalized, err := s.prepareCustomerInput(ctx, tenantID, in)
	if err != nil {
		return nil, err
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, _, err := s.writeCustomerTx(ctx, tx, tenantID, id, in, emailNormalized); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetCustomer(ctx, tenantID, id)
}

// CommitCustomerImport commits all prevalidated rows and the import summary atomically.
func (s *Store) CommitCustomerImport(ctx context.Context, tenantID string, inputs []CustomerInput, job CustomerImportJob) ([]CustomerUpsertResult, *CustomerImportJob, error) {
	type prepared struct {
		input           CustomerInput
		emailNormalized string
	}
	items := make([]prepared, 0, len(inputs))
	for _, input := range inputs {
		p, normalized, err := s.prepareCustomerInput(ctx, tenantID, input)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, prepared{input: p, emailNormalized: normalized})
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	type outcomeID struct{ id, outcome string }
	written := make([]outcomeID, 0, len(items))
	for _, item := range items {
		id, outcome, err := s.writeCustomerTx(ctx, tx, tenantID, "", item.input, item.emailNormalized)
		if err != nil {
			return nil, nil, err
		}
		written = append(written, outcomeID{id: id, outcome: outcome})
	}
	for _, item := range written {
		if item.outcome == "created" {
			job.CreatedRows++
		} else {
			job.UpdatedRows++
		}
	}
	if err := s.createCustomerImportJobTx(ctx, tx, tenantID, &job); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	results := make([]CustomerUpsertResult, 0, len(written))
	for _, item := range written {
		customer, err := s.GetCustomer(ctx, tenantID, item.id)
		if err != nil {
			return nil, nil, err
		}
		results = append(results, CustomerUpsertResult{Customer: customer, Outcome: item.outcome})
	}
	storedJob, err := s.GetCustomerImportJob(ctx, tenantID, job.ID)
	if err != nil {
		return nil, nil, err
	}
	return results, storedJob, nil
}

func (s *Store) writeCustomerTx(ctx context.Context, tx pgx.Tx, tenantID, forcedID string, in CustomerInput, emailNormalized string) (string, string, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	id := forcedID
	if id == "" {
		var byExternal, byEmail string
		if in.ExternalID != "" {
			err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT id FROM %s.customers WHERE tenant_id=$1 AND source=$2 AND external_id=$3`, schema), tenantID, in.Source, in.ExternalID).Scan(&byExternal)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return "", "", err
			}
		}
		if emailNormalized != "" {
			err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT id FROM %s.customers WHERE tenant_id=$1 AND email_normalized=$2`, schema), tenantID, emailNormalized).Scan(&byEmail)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return "", "", err
			}
		}
		if byExternal != "" && byEmail != "" && byExternal != byEmail {
			return "", "", ErrCustomerConflict
		}
		id = byExternal
		if id == "" {
			id = byEmail
		}
	}
	outcome := "updated"
	actor := auditctx.ActorID(ctx)
	metadata, _ := json.Marshal(in.Metadata)
	var emailArg, externalArg, tierArg any
	if emailNormalized != "" {
		emailArg = emailNormalized
	}
	if in.ExternalID != "" {
		externalArg = in.ExternalID
	}
	if in.TierID != "" {
		tierArg = in.TierID
	}
	var err error
	if id == "" {
		id = "cust_" + newStoreID()
		outcome = "created"
		_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customers
(id,tenant_id,email,email_normalized,phone,display_name,locale,tier_id,source,external_id,status,metadata,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$13)`, schema),
			id, tenantID, in.Email, emailArg, in.Phone, in.DisplayName, in.Locale, tierArg, in.Source, externalArg, in.Status, metadata, actor)
	} else {
		tag, execErr := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.customers SET
email=$3,email_normalized=$4,phone=$5,display_name=$6,locale=$7,tier_id=$8,source=$9,external_id=$10,
status=$11,metadata=$12,updated_by=$13,updated_at=now() WHERE tenant_id=$1 AND id=$2`, schema),
			tenantID, id, in.Email, emailArg, in.Phone, in.DisplayName, in.Locale, tierArg, in.Source, externalArg, in.Status, metadata, actor)
		err = execErr
		if err == nil && tag.RowsAffected() == 0 {
			return "", "", ErrCustomerNotFound
		}
	}
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return "", "", ErrCustomerConflict
		}
		return "", "", err
	}
	if _, err = tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.customer_group_members WHERE tenant_id=$1 AND customer_id=$2`, schema), tenantID, id); err != nil {
		return "", "", err
	}
	seen := map[string]bool{}
	for _, groupID := range in.GroupIDs {
		groupID = strings.TrimSpace(groupID)
		if groupID == "" || seen[groupID] {
			continue
		}
		seen[groupID] = true
		if _, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customer_group_members
(customer_id,group_id,tenant_id,created_by,updated_by) VALUES ($1,$2,$3,$4,$4)`, schema), id, groupID, tenantID, actor); err != nil {
			return "", "", err
		}
	}
	return id, outcome, nil
}

func (s *Store) scanCustomer(ctx context.Context, tenantID string, row pgx.Row) (*Customer, error) {
	var c Customer
	var emailNorm, tierID, externalID *string
	var metadata []byte
	err := row.Scan(&c.ID, &c.TenantID, &c.Email, &emailNorm, &c.Phone, &c.DisplayName, &c.Locale,
		&tierID, &c.Source, &externalID, &c.Status, &metadata, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if tierID != nil {
		c.TierID = *tierID
	}
	if externalID != nil {
		c.ExternalID = *externalID
	}
	c.Metadata = map[string]any{}
	_ = json.Unmarshal(metadata, &c.Metadata)
	c.GroupIDs = []string{}
	if s.pg != nil {
		schema := quoteIdent(s.cfg.PostgresSchema)
		rows, qerr := s.pg.Query(ctx, fmt.Sprintf(`SELECT group_id FROM %s.customer_group_members WHERE tenant_id=$1 AND customer_id=$2 ORDER BY group_id`, schema), tenantID, c.ID)
		if qerr != nil {
			return nil, qerr
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if qerr := rows.Scan(&id); qerr != nil {
				return nil, qerr
			}
			c.GroupIDs = append(c.GroupIDs, id)
		}
	}
	return &c, nil
}

func (s *Store) GetCustomer(ctx context.Context, tenantID, id string) (*Customer, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	c, err := s.scanCustomer(ctx, tenantID, s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,email,email_normalized,phone,display_name,locale,tier_id,source,external_id,status,metadata,created_at,updated_at FROM %s.customers WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}
	return c, err
}

func (s *Store) ListCustomers(ctx context.Context, tenantID string, f CustomerListFilter) ([]Customer, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 200 {
		f.Limit = 200
	}
	if f.Status != "" && f.Status != "active" && f.Status != "inactive" {
		return nil, fmt.Errorf("invalid status")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	q := strings.ToLower(strings.TrimSpace(f.Query))
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT id,tenant_id,email,email_normalized,phone,display_name,locale,tier_id,source,external_id,status,metadata,created_at,updated_at
FROM %s.customers WHERE tenant_id=$1
AND ($2='' OR status=$2) AND ($3='' OR tier_id=$3)
AND ($4='' OR lower(display_name) LIKE '%%'||$4||'%%' OR COALESCE(email_normalized,'') LIKE '%%'||$4||'%%' OR lower(phone) LIKE '%%'||$4||'%%' OR lower(COALESCE(external_id,'')) LIKE '%%'||$4||'%%')
ORDER BY updated_at DESC LIMIT $5`, schema), tenantID, f.Status, f.TierID, q, f.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Customer{}
	for rows.Next() {
		c, err := s.scanCustomer(ctx, tenantID, rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (s *Store) DeactivateCustomer(ctx context.Context, tenantID, id string) (*Customer, error) {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.customers SET status='inactive',updated_by=$3,updated_at=now() WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id, actor)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrCustomerNotFound
	}
	return s.GetCustomer(ctx, tenantID, id)
}

func (s *Store) CreateCustomerImportJob(ctx context.Context, tenantID string, job CustomerImportJob) (*CustomerImportJob, error) {
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := s.createCustomerImportJobTx(ctx, tx, tenantID, &job); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetCustomerImportJob(ctx, tenantID, job.ID)
}

func (s *Store) createCustomerImportJobTx(ctx context.Context, tx pgx.Tx, tenantID string, job *CustomerImportJob) error {
	if job.ID == "" {
		job.ID = "cimp_" + newStoreID()
	}
	actor := auditctx.ActorID(ctx)
	errJSON, _ := json.Marshal(job.Errors)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customer_import_jobs
(id,tenant_id,filename,mode,status,total_rows,created_rows,updated_rows,rejected_rows,errors,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$11)`, schema), job.ID, tenantID, job.Filename, job.Mode, job.Status, job.TotalRows, job.CreatedRows, job.UpdatedRows, job.RejectedRows, errJSON, actor)
	return err
}

func (s *Store) GetCustomerImportJob(ctx context.Context, tenantID, id string) (*CustomerImportJob, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var job CustomerImportJob
	var errJSON []byte
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,filename,mode,status,total_rows,created_rows,updated_rows,rejected_rows,errors,created_at,updated_at FROM %s.customer_import_jobs WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id).Scan(
		&job.ID, &job.TenantID, &job.Filename, &job.Mode, &job.Status, &job.TotalRows, &job.CreatedRows, &job.UpdatedRows, &job.RejectedRows, &errJSON, &job.CreatedAt, &job.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrImportNotFound
	}
	if err != nil {
		return nil, err
	}
	job.AcceptedRows = job.TotalRows - job.RejectedRows
	job.Errors = []map[string]any{}
	_ = json.Unmarshal(errJSON, &job.Errors)
	return &job, nil
}

func (s *Store) ListCustomerDomainRules(ctx context.Context, tenantID string) ([]CustomerDomainRule, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT id,tenant_id,domain,policy,default_tier_id,default_group_id,active,created_at,updated_at FROM %s.customer_domain_rules WHERE tenant_id=$1 ORDER BY domain`, schema), tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CustomerDomainRule{}
	for rows.Next() {
		rule, err := scanDomainRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rule)
	}
	return out, rows.Err()
}

func scanDomainRule(row pgx.Row) (*CustomerDomainRule, error) {
	var rule CustomerDomainRule
	var tierID, groupID *string
	err := row.Scan(&rule.ID, &rule.TenantID, &rule.Domain, &rule.Policy, &tierID, &groupID, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt)
	if tierID != nil {
		rule.DefaultTierID = *tierID
	}
	if groupID != nil {
		rule.DefaultGroupID = *groupID
	}
	return &rule, err
}

func (s *Store) GetCustomerDomainRule(ctx context.Context, tenantID, id string) (*CustomerDomainRule, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rule, err := scanDomainRule(s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,domain,policy,default_tier_id,default_group_id,active,created_at,updated_at FROM %s.customer_domain_rules WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDomainRuleNotFound
	}
	return rule, err
}

func (s *Store) FindCustomerDomainRule(ctx context.Context, tenantID, domain string) (*CustomerDomainRule, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rule, err := scanDomainRule(s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,domain,policy,default_tier_id,default_group_id,active,created_at,updated_at FROM %s.customer_domain_rules WHERE tenant_id=$1 AND domain=$2`, schema), tenantID, strings.ToLower(domain)))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDomainRuleNotFound
	}
	return rule, err
}

func (s *Store) validateDomainRuleInput(ctx context.Context, tenantID string, in CustomerDomainRuleInput) (CustomerDomainRuleInput, error) {
	domain, err := NormalizeCustomerDomain(in.Domain)
	if err != nil {
		return in, err
	}
	in.Domain = domain
	in.Policy = strings.ToLower(strings.TrimSpace(in.Policy))
	if in.Policy != "allow" && in.Policy != "deny" {
		return in, fmt.Errorf("policy must be allow or deny")
	}
	if err := s.validateCustomerAssignments(ctx, tenantID, in.DefaultTierID, []string{in.DefaultGroupID}); err != nil {
		return in, err
	}
	return in, nil
}

func (s *Store) CreateCustomerDomainRule(ctx context.Context, tenantID string, in CustomerDomainRuleInput) (*CustomerDomainRule, error) {
	in, err := s.validateDomainRuleInput(ctx, tenantID, in)
	if err != nil {
		return nil, err
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	id := "cdr_" + newStoreID()
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customer_domain_rules (id,tenant_id,domain,policy,default_tier_id,default_group_id,active,created_by,updated_by) VALUES ($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),$7,$8,$8)`, schema), id, tenantID, in.Domain, in.Policy, in.DefaultTierID, in.DefaultGroupID, active, actor)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, ErrDomainRuleTaken
		}
		return nil, err
	}
	return s.GetCustomerDomainRule(ctx, tenantID, id)
}

func (s *Store) UpdateCustomerDomainRule(ctx context.Context, tenantID, id string, in CustomerDomainRuleInput) (*CustomerDomainRule, error) {
	cur, err := s.GetCustomerDomainRule(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if in.Domain == "" {
		in.Domain = cur.Domain
	}
	if in.Policy == "" {
		in.Policy = cur.Policy
	}
	in, err = s.validateDomainRuleInput(ctx, tenantID, in)
	if err != nil {
		return nil, err
	}
	active := cur.Active
	if in.Active != nil {
		active = *in.Active
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.customer_domain_rules SET domain=$3,policy=$4,default_tier_id=NULLIF($5,''),default_group_id=NULLIF($6,''),active=$7,updated_by=$8,updated_at=now() WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id, in.Domain, in.Policy, in.DefaultTierID, in.DefaultGroupID, active, actor)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, ErrDomainRuleTaken
		}
		return nil, err
	}
	return s.GetCustomerDomainRule(ctx, tenantID, id)
}

func (s *Store) DeleteCustomerDomainRule(ctx context.Context, tenantID, id string) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.customer_domain_rules WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrDomainRuleNotFound
	}
	return nil
}
