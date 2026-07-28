package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrAvatarNotFound         = errors.New("avatar not found")
	ErrAvatarHasAssignments   = errors.New("avatar has active tenant assignments")
	ErrVoiceProviderNotFound  = errors.New("voice provider not found")
	ErrAssignmentNotFound     = errors.New("tenant avatar assignment not found")
	ErrMaxAIEmployeesExceeded = errors.New("max ai employees exceeded")
)

type Avatar struct {
	ID             string
	Slug           string
	Name           string
	Role           string
	Trait          string
	Color          string
	ImageURL       string
	Greeting       string
	Status         string
	OwnerTenantID  string // empty = platform catalog; set = tenant-owned library
	Flags          map[string]any
	Voices         []AvatarVoice
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AvatarVoice struct {
	ID              string
	AvatarID        string
	VoiceProviderID string
	VoiceID         string
	Voice           string
	Priority        int
	Status          string
}

type TenantAvatarAssignment struct {
	TenantID string
	AvatarID string
	Status   string
	Avatar   *Avatar
}

type WorkforceAgent struct {
	ID              string
	Name            string
	Role            string
	Trait           string
	Color           string
	Voice           string
	VoiceProviderID string
	VoiceID         string
	Image           string
	Greeting        string
	Popular         bool
	Robot           bool
	Skin            string
	Hair            string
}

func (s *Store) ensureAvatarsSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.ai_avatars (
  id text PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  name text NOT NULL,
  role text NOT NULL DEFAULT '',
  trait text NOT NULL DEFAULT '',
  color text NOT NULL DEFAULT '',
  image_url text NOT NULL DEFAULT '',
  greeting text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
  flags jsonb NOT NULL DEFAULT '{}'::jsonb,%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.ai_avatar_voices (
  id text PRIMARY KEY,
  avatar_id text NOT NULL REFERENCES %s.ai_avatars(id) ON DELETE CASCADE,
  voice_provider_id text NOT NULL REFERENCES %s.voice_providers(id),
  voice_id text NOT NULL,
  voice text NOT NULL,
  priority int NOT NULL DEFAULT 1,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),%s
)`, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_avatar_assignments (
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  avatar_id text NOT NULL REFERENCES %s.ai_avatars(id) ON DELETE CASCADE,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),%s,
  PRIMARY KEY (tenant_id, avatar_id)
)`, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS ai_avatar_voices_active_priority_idx
ON %s.ai_avatar_voices (avatar_id, priority) WHERE status = 'active'`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS tenant_avatar_assignments_active_idx
ON %s.tenant_avatar_assignments (tenant_id) WHERE status = 'active'`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	// Sprint 52: tenant-owned library avatars.
	if _, err := s.pg.Exec(ctx, fmt.Sprintf(`
ALTER TABLE %s.ai_avatars ADD COLUMN IF NOT EXISTS owner_tenant_id text REFERENCES %s.tenants(id) ON DELETE CASCADE`, schema, schema)); err != nil {
		return err
	}
	if _, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE INDEX IF NOT EXISTS ai_avatars_owner_tenant_idx
ON %s.ai_avatars (owner_tenant_id) WHERE owner_tenant_id IS NOT NULL`, schema)); err != nil {
		return err
	}
	return s.seedAvatars(ctx)
}

func (s *Store) seedAvatars(ctx context.Context) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	demoTenant := s.cfg.DemoTenantID
	if demoTenant == "" {
		demoTenant = "demo"
	}

	type seedAvatar struct {
		id, name, role, trait, color, imageURL, greeting, voice string
		flags                                                   string
	}
	seeds := []seedAvatar{
		{"ava", "Ava", "General Support", "Warm & Patient", "#008cff", "/images/ava.jpg",
			"Thank you for calling. I'm Ava from general support. How can I help you today?",
			"Aoede", `{"popular":true,"skin":"#f0bd9b","hair":"#5a3428"}`},
		{"max", "Max", "Billing Specialist", "Calm & Precise", "#0076ff", "/images/max.jpg",
			"Hi, this is Max from billing. I can help with invoices, payments, and account questions.",
			"Charon", `{"skin":"#e8ad88","hair":"#2d221f"}`},
		{"luna", "Luna", "Technical Support", "Clear & Helpful", "#b14dff", "/images/luna.jpg",
			"Hello, Luna here from technical support. Tell me what's going on and we'll troubleshoot it together.",
			"Kore", `{"skin":"#efc0a1","hair":"#7c52c8"}`},
		{"neo", "Neo", "Triage Bot", "Fast & Neutral", "#00a8ff", "/images/neo.jpg",
			"Neo triage online. Share your issue in one sentence and I'll route you to the right specialist.",
			"Puck", `{"robot":true}`},
	}
	for _, av := range seeds {
		_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.ai_avatars (id, slug, name, role, trait, color, image_url, greeting, status, flags)
VALUES ($1, $1, $2, $3, $4, $5, $6, $7, 'active', $8::jsonb)
ON CONFLICT (id) DO NOTHING`, schema),
			av.id, av.name, av.role, av.trait, av.color, av.imageURL, av.greeting, av.flags)
		if err != nil {
			return err
		}
		_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.ai_avatar_voices (id, avatar_id, voice_provider_id, voice_id, voice, priority, status)
VALUES ($1, $2, 'voice-gemini-live', 'gemini-2.5-flash-native-audio-latest', $3, 1, 'active')
ON CONFLICT (id) DO NOTHING`, schema),
			"avvoice_"+av.id+"_gemini", av.id, av.voice)
		if err != nil {
			return err
		}
		_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_avatar_assignments (tenant_id, avatar_id, status)
VALUES ($1, $2, 'active')
ON CONFLICT (tenant_id, avatar_id) DO NOTHING`, schema), demoTenant, av.id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListAvatars(ctx context.Context, status string) ([]Avatar, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	// Platform catalog list: only rows without a tenant owner.
	q := fmt.Sprintf(`
SELECT id, slug, name, role, trait, color, image_url, greeting, status, COALESCE(owner_tenant_id,''), flags, created_at, updated_at
FROM %s.ai_avatars WHERE owner_tenant_id IS NULL`, schema)
	args := []any{}
	if status != "" {
		q += ` AND status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY slug`
	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Avatar
	for rows.Next() {
		av, err := scanAvatar(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, av)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		voices, err := s.listAvatarVoices(ctx, out[i].ID, "")
		if err != nil {
			return nil, err
		}
		out[i].Voices = voices
	}
	return out, nil
}

func (s *Store) GetAvatar(ctx context.Context, id string) (*Avatar, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, slug, name, role, trait, color, image_url, greeting, status, COALESCE(owner_tenant_id,''), flags, created_at, updated_at
FROM %s.ai_avatars WHERE id = $1`, schema), id)
	av, err := scanAvatarRow(row)
	if err != nil {
		return nil, err
	}
	voices, err := s.listAvatarVoices(ctx, id, "")
	if err != nil {
		return nil, err
	}
	av.Voices = voices
	return av, nil
}

func (s *Store) CreateAvatar(ctx context.Context, avatar Avatar) (*Avatar, error) {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	flagsJSON, err := json.Marshal(avatar.Flags)
	if err != nil {
		return nil, err
	}
	if avatar.Flags == nil {
		flagsJSON = []byte("{}")
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var owner any
	if strings.TrimSpace(avatar.OwnerTenantID) != "" {
		owner = strings.TrimSpace(avatar.OwnerTenantID)
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.ai_avatars (id, slug, name, role, trait, color, image_url, greeting, status, owner_tenant_id, flags, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12, $12)`, schema),
		avatar.ID, avatar.Slug, avatar.Name, avatar.Role, avatar.Trait, avatar.Color,
		avatar.ImageURL, avatar.Greeting, avatar.Status, owner, string(flagsJSON), actor)
	if err != nil {
		return nil, err
	}
	for _, v := range avatar.Voices {
		if err := s.insertAvatarVoiceTx(ctx, tx, schema, avatar.ID, v, actor); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetAvatar(ctx, avatar.ID)
}

func (s *Store) UpdateAvatar(ctx context.Context, avatar Avatar) (*Avatar, error) {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	flagsJSON, err := json.Marshal(avatar.Flags)
	if err != nil {
		return nil, err
	}
	if avatar.Flags == nil {
		flagsJSON = []byte("{}")
	}
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.ai_avatars
SET slug = $2, name = $3, role = $4, trait = $5, color = $6, image_url = $7, greeting = $8, status = $9, flags = $10::jsonb, updated_by = $11
WHERE id = $1`, schema),
		avatar.ID, avatar.Slug, avatar.Name, avatar.Role, avatar.Trait, avatar.Color,
		avatar.ImageURL, avatar.Greeting, avatar.Status, string(flagsJSON), actor)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrAvatarNotFound
	}
	return s.GetAvatar(ctx, avatar.ID)
}

func (s *Store) ArchiveAvatar(ctx context.Context, id string) error {
	actor := auditctx.ActorID(ctx)
	has, err := s.AvatarHasActiveAssignments(ctx, id)
	if err != nil {
		return err
	}
	if has {
		return ErrAvatarHasAssignments
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.ai_avatars SET status = 'archived', updated_by = $2 WHERE id = $1`, schema), id, actor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAvatarNotFound
	}
	return nil
}

func (s *Store) ReplaceAvatarVoices(ctx context.Context, avatarID string, voices []AvatarVoice) error {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	if _, err := s.GetAvatar(ctx, avatarID); err != nil {
		return err
	}
	for _, v := range voices {
		if err := s.voiceProviderExists(ctx, v.VoiceProviderID); err != nil {
			return err
		}
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.ai_avatar_voices WHERE avatar_id = $1`, schema), avatarID)
	if err != nil {
		return err
	}
	for _, v := range voices {
		if err := s.insertAvatarVoiceTx(ctx, tx, schema, avatarID, v, actor); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) CountActiveTenantAssignments(ctx context.Context, tenantID string) (int, error) {
	if s == nil || s.pg == nil {
		return 0, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var n int
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s.tenant_avatar_assignments WHERE tenant_id = $1 AND status = 'active'`, schema), tenantID).Scan(&n)
	return n, err
}

func (s *Store) ListTenantAvatarAssignments(ctx context.Context, tenantID string) ([]TenantAvatarAssignment, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT ta.tenant_id, ta.avatar_id, ta.status
FROM %s.tenant_avatar_assignments ta
WHERE ta.tenant_id = $1
ORDER BY ta.avatar_id`, schema), tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TenantAvatarAssignment
	for rows.Next() {
		var a TenantAvatarAssignment
		if err := rows.Scan(&a.TenantID, &a.AvatarID, &a.Status); err != nil {
			return nil, err
		}
		av, err := s.GetAvatar(ctx, a.AvatarID)
		if err != nil {
			if errors.Is(err, ErrAvatarNotFound) {
				continue
			}
			return nil, err
		}
		a.Avatar = av
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) AssignAvatarToTenant(ctx context.Context, tenantID, avatarID string) (*TenantAvatarAssignment, error) {
	actor := auditctx.ActorID(ctx)
	exists, err := s.TenantExists(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTenantNotFound
	}
	av, err := s.GetAvatar(ctx, avatarID)
	if err != nil {
		return nil, err
	}
	if av.Status == "archived" {
		return nil, ErrAvatarNotFound
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_avatar_assignments (tenant_id, avatar_id, status, created_by, updated_by)
VALUES ($1, $2, 'active', $3, $3)
ON CONFLICT (tenant_id, avatar_id) DO UPDATE SET status = 'active', updated_by = $3`, schema),
		tenantID, avatarID, actor)
	if err != nil {
		return nil, err
	}
	return &TenantAvatarAssignment{TenantID: tenantID, AvatarID: avatarID, Status: "active", Avatar: av}, nil
}

func (s *Store) RevokeTenantAssignment(ctx context.Context, tenantID, avatarID string) error {
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_avatar_assignments SET status = 'disabled', updated_by = $3
WHERE tenant_id = $1 AND avatar_id = $2 AND status = 'active'`, schema), tenantID, avatarID, actor)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAssignmentNotFound
	}
	return nil
}

func (s *Store) ListWorkforceAgents(ctx context.Context, tenantID string) ([]WorkforceAgent, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT a.id, a.name, a.role, a.trait, a.color, a.image_url, a.greeting, a.flags,
       COALESCE(v.voice_provider_id, ''), COALESCE(v.voice_id, ''), COALESCE(v.voice, '')
FROM %s.tenant_avatar_assignments ta
JOIN %s.ai_avatars a ON a.id = ta.avatar_id
LEFT JOIN LATERAL (
  SELECT voice_provider_id, voice_id, voice
  FROM %s.ai_avatar_voices
  WHERE avatar_id = a.id AND status = 'active'
  ORDER BY priority ASC
  LIMIT 1
) v ON true
WHERE ta.tenant_id = $1 AND ta.status = 'active' AND a.status = 'active'
ORDER BY a.slug`, schema, schema, schema), tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WorkforceAgent
	for rows.Next() {
		var agent WorkforceAgent
		var flagsRaw []byte
		if err := rows.Scan(&agent.ID, &agent.Name, &agent.Role, &agent.Trait, &agent.Color,
			&agent.Image, &agent.Greeting, &flagsRaw,
			&agent.VoiceProviderID, &agent.VoiceID, &agent.Voice); err != nil {
			return nil, err
		}
		applyWorkforceFlags(&agent, flagsRaw)
		out = append(out, agent)
	}
	return out, rows.Err()
}

func (s *Store) HasTenantAvatarAssignments(ctx context.Context, tenantID string) bool {
	if s.pg == nil {
		return false
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var exists bool
	_ = s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT EXISTS(
  SELECT 1 FROM %s.tenant_avatar_assignments WHERE tenant_id = $1 AND status = 'active'
)`, schema), tenantID).Scan(&exists)
	return exists
}

func (s *Store) AvatarHasActiveAssignments(ctx context.Context, avatarID string) (bool, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var exists bool
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT EXISTS(
  SELECT 1 FROM %s.tenant_avatar_assignments WHERE avatar_id = $1 AND status = 'active'
)`, schema), avatarID).Scan(&exists)
	return exists, err
}

func (s *Store) listAvatarVoices(ctx context.Context, avatarID, status string) ([]AvatarVoice, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	q := fmt.Sprintf(`
SELECT id, avatar_id, voice_provider_id, voice_id, voice, priority, status
FROM %s.ai_avatar_voices WHERE avatar_id = $1`, schema)
	args := []any{avatarID}
	if status != "" {
		q += ` AND status = $2`
		args = append(args, status)
	}
	q += ` ORDER BY priority, id`
	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AvatarVoice
	for rows.Next() {
		var v AvatarVoice
		if err := rows.Scan(&v.ID, &v.AvatarID, &v.VoiceProviderID, &v.VoiceID, &v.Voice, &v.Priority, &v.Status); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) voiceProviderExists(ctx context.Context, id string) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var n int
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s.voice_providers WHERE id = $1`, schema), id).Scan(&n)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrVoiceProviderNotFound
	}
	return nil
}

func (s *Store) insertAvatarVoiceTx(ctx context.Context, tx pgx.Tx, schema, avatarID string, v AvatarVoice, actor string) error {
	if err := s.voiceProviderExists(ctx, v.VoiceProviderID); err != nil {
		return err
	}
	voiceID := v.ID
	if voiceID == "" {
		voiceID = fmt.Sprintf("avvoice_%s_%d", avatarID, v.Priority)
	}
	status := v.Status
	if status == "" {
		status = "active"
	}
	_, err := tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.ai_avatar_voices (id, avatar_id, voice_provider_id, voice_id, voice, priority, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)`, schema),
		voiceID, avatarID, v.VoiceProviderID, v.VoiceID, v.Voice, v.Priority, status, actor)
	return err
}

func scanAvatar(rows pgx.Rows) (Avatar, error) {
	var av Avatar
	var flagsRaw []byte
	if err := rows.Scan(&av.ID, &av.Slug, &av.Name, &av.Role, &av.Trait, &av.Color,
		&av.ImageURL, &av.Greeting, &av.Status, &av.OwnerTenantID, &flagsRaw, &av.CreatedAt, &av.UpdatedAt); err != nil {
		return Avatar{}, err
	}
	_ = json.Unmarshal(flagsRaw, &av.Flags)
	if av.Flags == nil {
		av.Flags = map[string]any{}
	}
	return av, nil
}

func scanAvatarRow(row pgx.Row) (*Avatar, error) {
	var av Avatar
	var flagsRaw []byte
	err := row.Scan(&av.ID, &av.Slug, &av.Name, &av.Role, &av.Trait, &av.Color,
		&av.ImageURL, &av.Greeting, &av.Status, &av.OwnerTenantID, &flagsRaw, &av.CreatedAt, &av.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAvatarNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(flagsRaw, &av.Flags)
	if av.Flags == nil {
		av.Flags = map[string]any{}
	}
	return &av, nil
}

// TenantLibraryAvatar is an avatar visible in the tenant library with assignment state.
type TenantLibraryAvatar struct {
	Avatar           Avatar
	AssignmentStatus string // active | disabled
}

// CreateTenantLibraryAvatar creates a tenant-owned avatar with inactive assignment.
// Does not enforce max_ai_employees (active cap applies only on activate).
// Optional: set avatar.Voices with a Gemini speaker name in Voice field.
func (s *Store) CreateTenantLibraryAvatar(ctx context.Context, tenantID string, avatar Avatar) (*TenantLibraryAvatar, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrTenantNotFound
	}
	exists, err := s.TenantExists(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTenantNotFound
	}
	avatar.OwnerTenantID = tenantID
	if avatar.Status == "" {
		avatar.Status = "active" // catalog row active; workforce gated by assignment
	}
	if avatar.ID == "" {
		avatar.ID = "av_" + newStoreID()
	}
	if avatar.Slug == "" {
		base := strings.ToLower(strings.TrimSpace(avatar.Name))
		base = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			if r == ' ' || r == '-' || r == '_' {
				return '-'
			}
			return -1
		}, base)
		if base == "" {
			base = "agent"
		}
		avatar.Slug = fmt.Sprintf("t-%s-%s", shortTenantFP(tenantID), base)
		if len(avatar.Slug) > 48 {
			avatar.Slug = avatar.Slug[:48]
		}
		avatar.Slug = avatar.Slug + "-" + newStoreID()[:6]
	}
	// Normalize / default Gemini speaker voice (AI Studio generate-speech names).
	avatar.Voices = normalizeTenantVoices(avatar.Voices)
	created, err := s.CreateAvatar(ctx, avatar)
	if err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_avatar_assignments (tenant_id, avatar_id, status, created_by, updated_by)
VALUES ($1, $2, 'disabled', $3, $3)
ON CONFLICT (tenant_id, avatar_id) DO UPDATE SET status = 'disabled', updated_by = $3`, schema),
		tenantID, created.ID, actor)
	if err != nil {
		return nil, err
	}
	return &TenantLibraryAvatar{Avatar: *created, AssignmentStatus: "disabled"}, nil
}

func shortTenantFP(tenantID string) string {
	var sum uint32
	for i := 0; i < len(tenantID); i++ {
		sum = sum*33 + uint32(tenantID[i])
	}
	return fmt.Sprintf("%02x", sum&0xff)
}

func normalizeTenantVoices(voices []AvatarVoice) []AvatarVoice {
	speaker := "Aoede"
	if len(voices) > 0 {
		if canon, ok := CanonicalGeminiSpeaker(voices[0].Voice); ok {
			speaker = canon
		} else if strings.TrimSpace(voices[0].Voice) != "" {
			// reject invalid later at handler; keep empty so create can fail closed
			speaker = strings.TrimSpace(voices[0].Voice)
		}
	}
	return []AvatarVoice{{
		VoiceProviderID: GeminiLiveProviderID,
		VoiceID:         GeminiLiveModelKey,
		Voice:           speaker,
		Priority:        1,
		Status:          "active",
	}}
}

// SetTenantAvatarSpeaker updates the primary Gemini speaker name for a tenant-owned avatar.
func (s *Store) SetTenantAvatarSpeaker(ctx context.Context, tenantID, avatarID, speaker string) error {
	canon, ok := CanonicalGeminiSpeaker(speaker)
	if !ok {
		return fmt.Errorf("invalid gemini speaker voice")
	}
	av, err := s.GetAvatar(ctx, avatarID)
	if err != nil {
		return err
	}
	if av.OwnerTenantID != tenantID {
		return ErrAvatarNotFound
	}
	// Ensure voice provider seed exists.
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, _ = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.voice_providers (id, provider, model_key, modality, status)
VALUES ('voice-gemini-live', 'google', 'gemini-2.5-flash-native-audio-latest', 'audio', 'active')
ON CONFLICT (id) DO NOTHING`, schema))

	voices := []AvatarVoice{{
		ID:              fmt.Sprintf("avvoice_%s_gemini", avatarID),
		VoiceProviderID: GeminiLiveProviderID,
		VoiceID:         GeminiLiveModelKey,
		Voice:           canon,
		Priority:        1,
		Status:          "active",
	}}
	return s.ReplaceAvatarVoices(ctx, avatarID, voices)
}

// PrimarySpeakerName returns the priority-1 active voice speaker name for an avatar.
func (s *Store) PrimarySpeakerName(ctx context.Context, avatarID string) string {
	voices, err := s.listAvatarVoices(ctx, avatarID, "active")
	if err != nil || len(voices) == 0 {
		return ""
	}
	// lowest priority first already ordered
	return voices[0].Voice
}

// ListTenantLibraryAvatars returns assignments for the tenant (library + active).
func (s *Store) ListTenantLibraryAvatars(ctx context.Context, tenantID string) ([]TenantLibraryAvatar, error) {
	assignments, err := s.ListTenantAvatarAssignments(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]TenantLibraryAvatar, 0, len(assignments))
	for _, a := range assignments {
		if a.Avatar == nil {
			continue
		}
		out = append(out, TenantLibraryAvatar{Avatar: *a.Avatar, AssignmentStatus: a.Status})
	}
	return out, nil
}

// GetTenantLibraryAvatar returns avatar if owned by or assigned to tenant.
func (s *Store) GetTenantLibraryAvatar(ctx context.Context, tenantID, avatarID string) (*TenantLibraryAvatar, error) {
	av, err := s.GetAvatar(ctx, avatarID)
	if err != nil {
		return nil, err
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var status string
	err = s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT status FROM %s.tenant_avatar_assignments WHERE tenant_id = $1 AND avatar_id = $2`, schema),
		tenantID, avatarID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		// Owned by tenant but missing assignment row — repair path not required; deny.
		if av.OwnerTenantID != tenantID {
			return nil, ErrAvatarNotFound
		}
		return nil, ErrAssignmentNotFound
	}
	if err != nil {
		return nil, err
	}
	if av.OwnerTenantID != "" && av.OwnerTenantID != tenantID {
		return nil, ErrAvatarNotFound
	}
	// Platform-catalog avatars may be assigned to the tenant.
	if av.OwnerTenantID == "" {
		// assigned is enough
	}
	return &TenantLibraryAvatar{Avatar: *av, AssignmentStatus: status}, nil
}

// UpdateTenantOwnedAvatar updates metadata only for tenant-owned avatars.
func (s *Store) UpdateTenantOwnedAvatar(ctx context.Context, tenantID string, avatar Avatar) (*Avatar, error) {
	existing, err := s.GetAvatar(ctx, avatar.ID)
	if err != nil {
		return nil, err
	}
	if existing.OwnerTenantID != tenantID {
		return nil, ErrAvatarNotFound
	}
	avatar.OwnerTenantID = tenantID
	// Preserve slug if not changing ownership identity.
	if avatar.Slug == "" {
		avatar.Slug = existing.Slug
	}
	return s.UpdateAvatar(ctx, avatar)
}

// ActivateTenantAvatar sets assignment active (caller must enforce quota).
func (s *Store) ActivateTenantAvatar(ctx context.Context, tenantID, avatarID string) (*TenantLibraryAvatar, error) {
	item, err := s.GetTenantLibraryAvatar(ctx, tenantID, avatarID)
	if err != nil {
		return nil, err
	}
	if item.Avatar.Status == "archived" {
		return nil, ErrAvatarNotFound
	}
	if item.AssignmentStatus == "active" {
		return item, nil
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_avatar_assignments SET status = 'active', updated_by = $3
WHERE tenant_id = $1 AND avatar_id = $2`, schema), tenantID, avatarID, actor)
	if err != nil {
		return nil, err
	}
	item.AssignmentStatus = "active"
	return item, nil
}

// DeactivateTenantAvatar sets assignment disabled.
func (s *Store) DeactivateTenantAvatar(ctx context.Context, tenantID, avatarID string) error {
	return s.RevokeTenantAssignment(ctx, tenantID, avatarID)
}

// ArchiveTenantOwnedAvatar disables assignment and archives the catalog row.
func (s *Store) ArchiveTenantOwnedAvatar(ctx context.Context, tenantID, avatarID string) error {
	av, err := s.GetAvatar(ctx, avatarID)
	if err != nil {
		return err
	}
	if av.OwnerTenantID != tenantID {
		return ErrAvatarNotFound
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	// Force-disable assignment even if active (archive frees workforce slot).
	_, _ = s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_avatar_assignments SET status = 'disabled', updated_by = $3
WHERE tenant_id = $1 AND avatar_id = $2`, schema), tenantID, avatarID, actor)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.ai_avatars SET status = 'archived', updated_by = $2 WHERE id = $1 AND owner_tenant_id = $3`, schema),
		avatarID, actor, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAvatarNotFound
	}
	return nil
}

func applyWorkforceFlags(agent *WorkforceAgent, flagsRaw []byte) {
	var flags map[string]any
	_ = json.Unmarshal(flagsRaw, &flags)
	if flags == nil {
		return
	}
	if v, ok := flags["popular"].(bool); ok {
		agent.Popular = v
	}
	if v, ok := flags["robot"].(bool); ok {
		agent.Robot = v
	}
	if v, ok := flags["skin"].(string); ok {
		agent.Skin = v
	}
	if v, ok := flags["hair"].(string); ok {
		agent.Hair = v
	}
}
