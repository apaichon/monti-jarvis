package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/libra/monti-jarvis/internal/leads"
)

var (
	ErrLeadNotFound = errors.New("lead not found")
	ErrLeadInvalid  = errors.New("invalid lead")
)

// MarketingLead is a sales lead captured from product-web.
type MarketingLead struct {
	ID                string     `json:"id"`
	Kind              string     `json:"kind"`
	Status            string     `json:"status"`
	Email             string     `json:"email"`
	FullName          string     `json:"full_name,omitempty"`
	CompanyName       string     `json:"company_name,omitempty"`
	Phone             string     `json:"phone,omitempty"`
	UseCase           string     `json:"use_case,omitempty"`
	PreferredChannel  string     `json:"preferred_channel,omitempty"`
	Language          string     `json:"language"`
	ConsentMarketing  bool       `json:"consent_marketing"`
	ConsentContact    bool       `json:"consent_contact"`
	ConsentAt         *time.Time `json:"consent_at,omitempty"`
	UTMSource         string     `json:"utm_source,omitempty"`
	UTMMedium         string     `json:"utm_medium,omitempty"`
	UTMCampaign       string     `json:"utm_campaign,omitempty"`
	UTMContent        string     `json:"utm_content,omitempty"`
	UTMTerm           string     `json:"utm_term,omitempty"`
	ReferralCode      string     `json:"referral_code,omitempty"`
	LandingPath       string     `json:"landing_path,omitempty"`
	PackageInterestID string     `json:"package_interest_id,omitempty"`
	DedupeKey         string     `json:"dedupe_key,omitempty"`
	AssignedTo        string     `json:"assigned_to,omitempty"`
	ConvertedTenantID string     `json:"converted_tenant_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	CreatedBy         string     `json:"created_by,omitempty"`
	UpdatedBy         string     `json:"updated_by,omitempty"`
}

// MarketingLeadNote is a platform follow-up note.
type MarketingLeadNote struct {
	ID        string    `json:"id"`
	LeadID    string    `json:"lead_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by,omitempty"`
}

// MarketingLeadEvent is a status transition history row.
type MarketingLeadEvent struct {
	ID         string    `json:"id"`
	LeadID     string    `json:"lead_id"`
	FromStatus string    `json:"from_status,omitempty"`
	ToStatus   string    `json:"to_status"`
	Actor      string    `json:"actor"`
	CreatedAt  time.Time `json:"created_at"`
}

// LeadListFilters filters platform lead list.
type LeadListFilters struct {
	Status string
	Kind   string
	Source string // utm_source
	Q      string // email/company substring
	Limit  int
	Offset int
}

// CreateLeadResult is returned from CreateLead.
type CreateLeadResult struct {
	Lead    MarketingLead
	Deduped bool
}

// FunnelEventInput is stored from public funnel beacons.
type FunnelEventInput struct {
	EventName    string
	PagePath     string
	CTAID        string
	SessionKey   string
	UTMSource    string
	UTMMedium    string
	UTMCampaign  string
	UTMContent   string
	UTMTerm      string
	ReferralCode string
	ClientIPHash string
}

func (s *Store) ensureLeadsSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.marketing_leads (
  id text PRIMARY KEY,
  kind text NOT NULL CHECK (kind IN ('contact', 'book_demo', 'newsletter')),
  status text NOT NULL DEFAULT 'new' CHECK (status IN (
    'new', 'contacted', 'demo_scheduled', 'demo_completed', 'qualified',
    'registered', 'kyc_pending', 'package_selected', 'paid', 'converted',
    'lost', 'unsubscribed'
  )),
  email text NOT NULL,
  full_name text NOT NULL DEFAULT '',
  company_name text NOT NULL DEFAULT '',
  phone text NOT NULL DEFAULT '',
  use_case text NOT NULL DEFAULT '',
  preferred_channel text NOT NULL DEFAULT '' CHECK (preferred_channel IN ('', 'email', 'phone', 'line', 'other')),
  language text NOT NULL DEFAULT 'en' CHECK (language IN ('en', 'th')),
  consent_marketing boolean NOT NULL DEFAULT false,
  consent_contact boolean NOT NULL DEFAULT false,
  consent_at timestamptz,
  utm_source text NOT NULL DEFAULT '',
  utm_medium text NOT NULL DEFAULT '',
  utm_campaign text NOT NULL DEFAULT '',
  utm_content text NOT NULL DEFAULT '',
  utm_term text NOT NULL DEFAULT '',
  referral_code text NOT NULL DEFAULT '',
  landing_path text NOT NULL DEFAULT '',
  package_interest_id text NOT NULL DEFAULT '',
  dedupe_key text NOT NULL DEFAULT '',
  assigned_to text NOT NULL DEFAULT '',
  converted_tenant_id text REFERENCES %s.tenants(id) ON DELETE SET NULL,%s,
  UNIQUE (kind, email)
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS marketing_leads_status_created_idx
ON %s.marketing_leads (status, created_at DESC)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS marketing_leads_kind_created_idx
ON %s.marketing_leads (kind, created_at DESC)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS marketing_leads_utm_source_idx
ON %s.marketing_leads (utm_source) WHERE utm_source <> ''`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.marketing_lead_notes (
  id text PRIMARY KEY,
  lead_id text NOT NULL REFERENCES %s.marketing_leads(id) ON DELETE CASCADE,
  body text NOT NULL,%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS marketing_lead_notes_lead_idx
ON %s.marketing_lead_notes (lead_id, created_at ASC)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.marketing_lead_events (
  id text PRIMARY KEY,
  lead_id text NOT NULL REFERENCES %s.marketing_leads(id) ON DELETE CASCADE,
  from_status text,
  to_status text NOT NULL,
  actor text NOT NULL DEFAULT 'system',
  created_at timestamptz NOT NULL DEFAULT now()
)`, schema, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS marketing_lead_events_lead_idx
ON %s.marketing_lead_events (lead_id, created_at ASC)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.funnel_events (
  id text PRIMARY KEY,
  event_name text NOT NULL,
  page_path text NOT NULL DEFAULT '',
  cta_id text NOT NULL DEFAULT '',
  utm_source text NOT NULL DEFAULT '',
  utm_medium text NOT NULL DEFAULT '',
  utm_campaign text NOT NULL DEFAULT '',
  utm_content text NOT NULL DEFAULT '',
  utm_term text NOT NULL DEFAULT '',
  referral_code text NOT NULL DEFAULT '',
  session_key text NOT NULL DEFAULT '',
  client_ip_hash text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS funnel_events_name_created_idx
ON %s.funnel_events (event_name, created_at DESC)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS funnel_events_created_idx
ON %s.funnel_events (created_at DESC)`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("leads schema: %w", err)
		}
	}
	return nil
}

// CreateLead inserts a lead or returns the existing row for (kind, email) within/via unique constraint.
func (s *Store) CreateLead(ctx context.Context, in leads.LeadInput, dedupeWindowHours int) (CreateLeadResult, error) {
	if s == nil || s.pg == nil {
		return CreateLeadResult{}, fmt.Errorf("postgres is not available")
	}
	in = leads.NormalizeLead(in)
	if err := leads.ValidateLead(in); err != nil {
		return CreateLeadResult{}, err
	}
	if dedupeWindowHours <= 0 {
		dedupeWindowHours = 24
	}

	schema := quoteIdent(s.cfg.PostgresSchema)
	existing, err := s.getLeadByKindEmail(ctx, in.Kind, in.Email)
	if err == nil {
		// Always treat unique (kind,email) as dedupe hit; window documented for analytics.
		_ = dedupeWindowHours
		return CreateLeadResult{Lead: existing, Deduped: true}, nil
	}
	if !errors.Is(err, ErrLeadNotFound) {
		return CreateLeadResult{}, err
	}

	id := "lead_" + newStoreID()
	dedupeKey := leads.DedupeKey(in.Kind, in.Email)
	actor := auditctx.ActorID(ctx)
	now := time.Now().UTC()
	consentAt := now

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return CreateLeadResult{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.marketing_leads (
  id, kind, status, email, full_name, company_name, phone, use_case,
  preferred_channel, language, consent_marketing, consent_contact, consent_at,
  utm_source, utm_medium, utm_campaign, utm_content, utm_term,
  referral_code, landing_path, package_interest_id, dedupe_key,
  created_by, updated_by
) VALUES (
  $1,$2,'new',$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$22
)`, schema),
		id, in.Kind, in.Email, in.FullName, in.CompanyName, in.Phone, in.UseCase,
		in.PreferredChannel, in.Language, in.ConsentMarketing, in.ConsentContact, consentAt,
		in.UTMSource, in.UTMMedium, in.UTMCampaign, in.UTMContent, in.UTMTerm,
		in.ReferralCode, in.LandingPath, in.PackageInterestID, dedupeKey, actor)
	if err != nil {
		// Race: unique (kind,email) — return existing.
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "unique") {
			existing, lookupErr := s.getLeadByKindEmail(ctx, in.Kind, in.Email)
			if lookupErr == nil {
				return CreateLeadResult{Lead: existing, Deduped: true}, nil
			}
		}
		return CreateLeadResult{}, err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.marketing_lead_events (id, lead_id, from_status, to_status, actor)
VALUES ($1,$2,NULL,'new',$3)`, schema),
		"le_"+newStoreID(), id, "system")
	if err != nil {
		return CreateLeadResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateLeadResult{}, err
	}

	lead, err := s.GetLead(ctx, id)
	if err != nil {
		return CreateLeadResult{}, err
	}
	return CreateLeadResult{Lead: lead, Deduped: false}, nil
}

// ListLeads returns filtered marketing leads for platform sales.
func (s *Store) ListLeads(ctx context.Context, f LeadListFilters) ([]MarketingLead, int, error) {
	if s == nil || s.pg == nil {
		return nil, 0, fmt.Errorf("postgres is not available")
	}
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 200 {
		f.Limit = 200
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	schema := quoteIdent(s.cfg.PostgresSchema)
	where := []string{"1=1"}
	args := []any{}
	argN := 1
	if status := strings.TrimSpace(f.Status); status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, status)
		argN++
	}
	if kind := strings.TrimSpace(f.Kind); kind != "" {
		where = append(where, fmt.Sprintf("kind = $%d", argN))
		args = append(args, kind)
		argN++
	}
	if source := strings.TrimSpace(f.Source); source != "" {
		where = append(where, fmt.Sprintf("utm_source = $%d", argN))
		args = append(args, source)
		argN++
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		where = append(where, fmt.Sprintf("(email ILIKE $%d OR company_name ILIKE $%d OR full_name ILIKE $%d)", argN, argN, argN))
		args = append(args, "%"+q+"%")
		argN++
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.marketing_leads WHERE %s`, schema, clause), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT %s FROM %s.marketing_leads
WHERE %s
ORDER BY created_at DESC
LIMIT $%d OFFSET $%d`, leadSelectCols, schema, clause, argN, argN+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanLeads(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetLead returns a lead by id.
func (s *Store) GetLead(ctx context.Context, id string) (MarketingLead, error) {
	if s == nil || s.pg == nil {
		return MarketingLead{}, fmt.Errorf("postgres is not available")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return MarketingLead{}, ErrLeadNotFound
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT %s FROM %s.marketing_leads WHERE id = $1`, leadSelectCols, schema), id)
	lead, err := scanLead(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return MarketingLead{}, ErrLeadNotFound
	}
	return lead, err
}

// GetLeadDetail returns lead, notes, and status history.
func (s *Store) GetLeadDetail(ctx context.Context, id string) (MarketingLead, []MarketingLeadNote, []MarketingLeadEvent, error) {
	lead, err := s.GetLead(ctx, id)
	if err != nil {
		return MarketingLead{}, nil, nil, err
	}
	notes, err := s.listLeadNotes(ctx, id)
	if err != nil {
		return MarketingLead{}, nil, nil, err
	}
	events, err := s.listLeadEvents(ctx, id)
	if err != nil {
		return MarketingLead{}, nil, nil, err
	}
	return lead, notes, events, nil
}

// UpdateLeadStatus patches status and/or assignment; records history on status change.
func (s *Store) UpdateLeadStatus(ctx context.Context, id, status, assignedTo string) (MarketingLead, error) {
	if s == nil || s.pg == nil {
		return MarketingLead{}, fmt.Errorf("postgres is not available")
	}
	id = strings.TrimSpace(id)
	status = strings.ToLower(strings.TrimSpace(status))
	assignedTo = strings.TrimSpace(assignedTo)

	existing, err := s.GetLead(ctx, id)
	if err != nil {
		return MarketingLead{}, err
	}

	newStatus := existing.Status
	if status != "" {
		if !leads.ValidLeadStatus(status) {
			return MarketingLead{}, ErrLeadInvalid
		}
		newStatus = status
	}
	newAssigned := existing.AssignedTo
	if assignedTo != "" || status != "" {
		// Allow clearing assignment only when explicitly set via sentinel? Keep provided value if non-empty;
		// if caller passes assigned_to as empty string and only status changes, keep previous.
		if assignedTo != "" {
			newAssigned = assignedTo
		}
	}
	// Support explicit unassign via "-" sentinel is overkill; accept assigned_to as provided when key present
	// at handler layer. Here: if assignedTo is the only update and empty, leave as-is unless status empty too.

	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return MarketingLead{}, err
	}
	defer tx.Rollback(ctx)

	// When assignedTo is explicitly provided from handler (including empty to clear), use UpdateLeadFields style.
	// This method receives assignedTo; empty means "do not change" unless we use a pointer. Handlers will
	// pass existing or new value.
	tag, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.marketing_leads
SET status = $2, assigned_to = $3, updated_by = $4
WHERE id = $1`, schema), id, newStatus, newAssigned, actor)
	if err != nil {
		return MarketingLead{}, err
	}
	if tag.RowsAffected() == 0 {
		return MarketingLead{}, ErrLeadNotFound
	}

	if newStatus != existing.Status {
		from := existing.Status
		_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.marketing_lead_events (id, lead_id, from_status, to_status, actor)
VALUES ($1,$2,$3,$4,$5)`, schema),
			"le_"+newStoreID(), id, from, newStatus, actor)
		if err != nil {
			return MarketingLead{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return MarketingLead{}, err
	}
	return s.GetLead(ctx, id)
}

// PatchLead updates status and/or assigned_to; nil pointer means leave unchanged.
func (s *Store) PatchLead(ctx context.Context, id string, status *string, assignedTo *string) (MarketingLead, error) {
	if s == nil || s.pg == nil {
		return MarketingLead{}, fmt.Errorf("postgres is not available")
	}
	existing, err := s.GetLead(ctx, id)
	if err != nil {
		return MarketingLead{}, err
	}
	newStatus := existing.Status
	if status != nil {
		st := strings.ToLower(strings.TrimSpace(*status))
		if st == "" || !leads.ValidLeadStatus(st) {
			return MarketingLead{}, ErrLeadInvalid
		}
		newStatus = st
	}
	newAssigned := existing.AssignedTo
	if assignedTo != nil {
		newAssigned = strings.TrimSpace(*assignedTo)
	}

	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return MarketingLead{}, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.marketing_leads
SET status = $2, assigned_to = $3, updated_by = $4
WHERE id = $1`, schema), id, newStatus, newAssigned, actor)
	if err != nil {
		return MarketingLead{}, err
	}
	if tag.RowsAffected() == 0 {
		return MarketingLead{}, ErrLeadNotFound
	}
	if newStatus != existing.Status {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.marketing_lead_events (id, lead_id, from_status, to_status, actor)
VALUES ($1,$2,$3,$4,$5)`, schema),
			"le_"+newStoreID(), id, existing.Status, newStatus, actor)
		if err != nil {
			return MarketingLead{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return MarketingLead{}, err
	}
	return s.GetLead(ctx, id)
}

// AddLeadNote appends a note to a lead.
func (s *Store) AddLeadNote(ctx context.Context, leadID, body string) (MarketingLeadNote, error) {
	if s == nil || s.pg == nil {
		return MarketingLeadNote{}, fmt.Errorf("postgres is not available")
	}
	if err := leads.ValidateNoteBody(body); err != nil {
		return MarketingLeadNote{}, err
	}
	body = strings.TrimSpace(body)
	if _, err := s.GetLead(ctx, leadID); err != nil {
		return MarketingLeadNote{}, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	id := "lnote_" + newStoreID()
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.marketing_lead_notes (id, lead_id, body, created_by, updated_by)
VALUES ($1,$2,$3,$4,$4)`, schema), id, leadID, body, actor)
	if err != nil {
		return MarketingLeadNote{}, err
	}
	var note MarketingLeadNote
	err = s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, lead_id, body, created_at, created_by FROM %s.marketing_lead_notes WHERE id = $1`, schema), id).
		Scan(&note.ID, &note.LeadID, &note.Body, &note.CreatedAt, &note.CreatedBy)
	return note, err
}

// InsertFunnelEvent stores a coarse funnel beacon (no PII body).
func (s *Store) InsertFunnelEvent(ctx context.Context, in FunnelEventInput) (string, error) {
	if s == nil || s.pg == nil {
		return "", fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	id := "fe_" + newStoreID()
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.funnel_events (
  id, event_name, page_path, cta_id, utm_source, utm_medium, utm_campaign,
  utm_content, utm_term, referral_code, session_key, client_ip_hash
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, schema),
		id, in.EventName, in.PagePath, in.CTAID, in.UTMSource, in.UTMMedium, in.UTMCampaign,
		in.UTMContent, in.UTMTerm, in.ReferralCode, in.SessionKey, in.ClientIPHash)
	if err != nil {
		return "", err
	}
	return id, nil
}

// ListPublicPackages returns active packages projected for marketing pricing.
// Includes both self-serve (Shared Cloud) and quote (Dedicated) catalogs.
func (s *Store) ListPublicPackages(ctx context.Context) ([]leads.PublicPackage, error) {
	pkgs, err := s.ListPackages(ctx, "active")
	if err != nil {
		return nil, err
	}
	src := make([]leads.PackageSource, 0, len(pkgs))
	for _, p := range pkgs {
		src = append(src, leads.PackageSource{
			ID: p.ID, Slug: p.Slug, Name: p.Name, Description: p.Description,
			PriceCents: p.PriceCents, Currency: p.Currency, BillingPeriod: p.BillingPeriod,
			Status: p.Status, Rules: p.Rules,
			PurchaseMode: PackagePurchaseMode(p),
			Deployment:   PackageDeployment(p),
		})
	}
	out := leads.ProjectPublicPackages(src)
	// Stable commercial order: shared by price, then dedicated by price.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Deployment != out[j].Deployment {
			return out[i].Deployment == "shared_cloud"
		}
		if out[i].PriceCurrency != out[j].PriceCurrency {
			return out[i].PriceCurrency < out[j].PriceCurrency
		}
		return out[i].PriceAmount < out[j].PriceAmount
	})
	return out, nil
}

func (s *Store) getLeadByKindEmail(ctx context.Context, kind, email string) (MarketingLead, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s FROM %s.marketing_leads WHERE kind = $1 AND email = $2`, leadSelectCols, schema), kind, email)
	lead, err := scanLead(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return MarketingLead{}, ErrLeadNotFound
	}
	return lead, err
}

func (s *Store) listLeadNotes(ctx context.Context, leadID string) ([]MarketingLeadNote, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT id, lead_id, body, created_at, created_by FROM %s.marketing_lead_notes
WHERE lead_id = $1 ORDER BY created_at ASC`, schema), leadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MarketingLeadNote
	for rows.Next() {
		var n MarketingLeadNote
		if err := rows.Scan(&n.ID, &n.LeadID, &n.Body, &n.CreatedAt, &n.CreatedBy); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) listLeadEvents(ctx context.Context, leadID string) ([]MarketingLeadEvent, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT id, lead_id, COALESCE(from_status,''), to_status, actor, created_at
FROM %s.marketing_lead_events WHERE lead_id = $1 ORDER BY created_at ASC`, schema), leadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MarketingLeadEvent
	for rows.Next() {
		var e MarketingLeadEvent
		if err := rows.Scan(&e.ID, &e.LeadID, &e.FromStatus, &e.ToStatus, &e.Actor, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

const leadSelectCols = `id, kind, status, email, full_name, company_name, phone, use_case,
preferred_channel, language, consent_marketing, consent_contact, consent_at,
utm_source, utm_medium, utm_campaign, utm_content, utm_term, referral_code,
landing_path, package_interest_id, dedupe_key, assigned_to, COALESCE(converted_tenant_id,''),
created_at, updated_at, created_by, updated_by`

type leadScanner interface {
	Scan(dest ...any) error
}

func scanLead(row leadScanner) (MarketingLead, error) {
	var l MarketingLead
	var consentAt *time.Time
	err := row.Scan(
		&l.ID, &l.Kind, &l.Status, &l.Email, &l.FullName, &l.CompanyName, &l.Phone, &l.UseCase,
		&l.PreferredChannel, &l.Language, &l.ConsentMarketing, &l.ConsentContact, &consentAt,
		&l.UTMSource, &l.UTMMedium, &l.UTMCampaign, &l.UTMContent, &l.UTMTerm, &l.ReferralCode,
		&l.LandingPath, &l.PackageInterestID, &l.DedupeKey, &l.AssignedTo, &l.ConvertedTenantID,
		&l.CreatedAt, &l.UpdatedAt, &l.CreatedBy, &l.UpdatedBy,
	)
	if err != nil {
		return MarketingLead{}, err
	}
	l.ConsentAt = consentAt
	return l, nil
}

func scanLeads(rows pgx.Rows) ([]MarketingLead, error) {
	var out []MarketingLead
	for rows.Next() {
		l, err := scanLead(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
