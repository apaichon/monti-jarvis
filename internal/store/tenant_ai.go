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
	"github.com/libra/monti-jarvis/internal/secretbox"
)

var (
	ErrTenantAIConfigNotFound    = errors.New("tenant ai config not found")
	ErrTenantSecretNotConfigured = errors.New("tenant secret encryption is not configured")
	ErrTenantSecretInvalid       = errors.New("tenant secret is invalid")
	ErrTenantPromptInvalid       = errors.New("tenant prompt is invalid")
	ErrTenantToolInvalid         = errors.New("tenant tool is invalid")
	ErrTenantToolInUse           = errors.New("tenant tool is in use")
	ErrTenantSkillInvalid        = errors.New("tenant skill is invalid")
)

type TenantAIConfig struct {
	TenantID     string     `json:"tenant_id"`
	Configured   bool       `json:"configured"`
	KeyLast4     string     `json:"last4,omitempty"`
	KeyVersion   string     `json:"key_version,omitempty"`
	KeyUpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type TenantAgentConfig struct {
	TenantID     string    `json:"tenant_id"`
	AgentID      string    `json:"agent_id"`
	SystemPrompt string    `json:"system_prompt"`
	Enabled      bool      `json:"enabled"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type TenantCallTool struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenant_id"`
	ToolKey     string         `json:"tool_key"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	HandlerKey  string         `json:"handler_key"`
	InputSchema map[string]any `json:"input_schema"`
	Enabled     bool           `json:"enabled"`
	UpdatedAt   time.Time      `json:"updated_at,omitempty"`
}

type TenantSkill struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Prompt    string    `json:"prompt"`
	Enabled   bool      `json:"enabled"`
	ToolIDs   []string  `json:"tool_ids,omitempty"`
	AgentIDs  []string  `json:"agent_ids,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// TenantAgentPrompt returns only the tenant-owned prompt fragments for an
// assigned agent. The immutable platform prompt remains in the orchestrator.
func (s *Store) TenantAgentPrompt(ctx context.Context, tenantID, agentID string) (string, error) {
	if s.pg == nil {
		return "", fmt.Errorf("postgres is not available")
	}
	base, err := s.GetTenantAgentConfig(ctx, tenantID, agentID)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, 3)
	if base.Enabled && strings.TrimSpace(base.SystemPrompt) != "" {
		parts = append(parts, "Tenant agent instructions:\n"+strings.TrimSpace(base.SystemPrompt))
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT sk.name, sk.prompt FROM %s.tenant_skills sk JOIN %s.tenant_agent_skills a ON a.tenant_id=sk.tenant_id AND a.skill_id=sk.id WHERE sk.tenant_id=$1 AND a.agent_id=$2 AND sk.enabled=true ORDER BY sk.name`, schema, schema), tenantID, agentID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var name, prompt string
		if err := rows.Scan(&name, &prompt); err != nil {
			return "", err
		}
		if strings.TrimSpace(prompt) != "" {
			parts = append(parts, "Tenant skill: "+strings.TrimSpace(name)+"\n"+strings.TrimSpace(prompt))
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return strings.Join(parts, "\n\n"), nil
}

func (s *Store) ensureTenantAISchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
ALTER TABLE %s.tenant_embed_configs ADD COLUMN IF NOT EXISTS auth_required boolean NOT NULL DEFAULT false;
CREATE TABLE IF NOT EXISTS %s.tenant_ai_configs (
  tenant_id text PRIMARY KEY REFERENCES %s.tenants(id) ON DELETE CASCADE,
  gemini_key_ciphertext bytea,
  gemini_key_nonce bytea,
  gemini_key_version text,
  gemini_key_last4 text,
  gemini_key_updated_at timestamptz,
  %s
);
CREATE TABLE IF NOT EXISTS %s.tenant_agent_configs (
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  agent_id text NOT NULL,
  system_prompt text NOT NULL DEFAULT '',
  enabled boolean NOT NULL DEFAULT true,
  %s,
  PRIMARY KEY (tenant_id, agent_id)
);
CREATE TABLE IF NOT EXISTS %s.tenant_call_tools (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  tool_key text NOT NULL,
  display_name text NOT NULL,
  description text NOT NULL,
  handler_key text NOT NULL,
  input_schema jsonb NOT NULL,
  enabled boolean NOT NULL DEFAULT false,
  %s,
  UNIQUE (tenant_id, tool_key)
);
CREATE TABLE IF NOT EXISTS %s.tenant_skills (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  slug text NOT NULL,
  name text NOT NULL,
  prompt text NOT NULL DEFAULT '',
  enabled boolean NOT NULL DEFAULT true,
  %s,
  UNIQUE (tenant_id, slug)
);
CREATE TABLE IF NOT EXISTS %s.tenant_skill_tools (
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  skill_id text NOT NULL,
  tool_id text NOT NULL,
  %s,
  PRIMARY KEY (tenant_id, skill_id, tool_id)
);
CREATE TABLE IF NOT EXISTS %s.tenant_agent_skills (
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  agent_id text NOT NULL,
  skill_id text NOT NULL,
  %s,
  PRIMARY KEY (tenant_id, agent_id, skill_id)
);
CREATE INDEX IF NOT EXISTS tenant_agent_configs_agent_idx ON %s.tenant_agent_configs (tenant_id, agent_id);
CREATE INDEX IF NOT EXISTS tenant_call_tools_tenant_idx ON %s.tenant_call_tools (tenant_id, enabled);
CREATE INDEX IF NOT EXISTS tenant_skills_tenant_idx ON %s.tenant_skills (tenant_id, enabled);`,
		schema, schema, schema, auditColumnsDDL, schema, schema, auditColumnsDDL,
		schema, schema, auditColumnsDDL, schema, schema, auditColumnsDDL,
		schema, schema, auditColumnsDDL, schema, schema, auditColumnsDDL,
		schema, schema, schema),
	)
	return err
}

func (s *Store) GetTenantAIConfig(ctx context.Context, tenantID string) (TenantAIConfig, error) {
	if s.pg == nil {
		return TenantAIConfig{}, fmt.Errorf("postgres is not available")
	}
	var out TenantAIConfig
	var ciphertext, nonce []byte
	schema := quoteIdent(s.cfg.PostgresSchema)
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT tenant_id, gemini_key_ciphertext, gemini_key_nonce, gemini_key_version, gemini_key_last4, gemini_key_updated_at FROM %s.tenant_ai_configs WHERE tenant_id=$1`, schema), tenantID).Scan(
		&out.TenantID, &ciphertext, &nonce, &out.KeyVersion, &out.KeyLast4, &out.KeyUpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantAIConfig{TenantID: tenantID}, nil
	}
	if err != nil {
		return TenantAIConfig{}, err
	}
	out.Configured = len(ciphertext) > 0 && len(nonce) > 0
	return out, nil
}

func (s *Store) PutTenantGeminiKey(ctx context.Context, tenantID, value string) (TenantAIConfig, error) {
	value = strings.TrimSpace(value)
	if len(value) < 20 || len(value) > 512 {
		return TenantAIConfig{}, ErrTenantSecretInvalid
	}
	key, err := secretbox.ParseKey(s.cfg.TenantSecretEncryptionKey)
	if err != nil {
		return TenantAIConfig{}, ErrTenantSecretNotConfigured
	}
	ciphertext, nonce, err := secretbox.Encrypt(key, []byte(value))
	if err != nil {
		return TenantAIConfig{}, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.tenant_ai_configs (tenant_id, gemini_key_ciphertext, gemini_key_nonce, gemini_key_version, gemini_key_last4, gemini_key_updated_at, created_by, updated_by) VALUES ($1,$2,$3,$4,$5,now(),$6,$6) ON CONFLICT (tenant_id) DO UPDATE SET gemini_key_ciphertext=EXCLUDED.gemini_key_ciphertext, gemini_key_nonce=EXCLUDED.gemini_key_nonce, gemini_key_version=EXCLUDED.gemini_key_version, gemini_key_last4=EXCLUDED.gemini_key_last4, gemini_key_updated_at=now(), updated_by=EXCLUDED.updated_by`, schema), tenantID, ciphertext, nonce, s.cfg.TenantSecretKeyVersion, secretbox.Last4(value), actor)
	if err != nil {
		return TenantAIConfig{}, err
	}
	return s.GetTenantAIConfig(ctx, tenantID)
}

func (s *Store) DeleteTenantGeminiKey(ctx context.Context, tenantID string) error {
	if s.pg == nil {
		return fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.tenant_ai_configs WHERE tenant_id=$1`, schema), tenantID)
	return err
}

func (s *Store) TenantGeminiKey(ctx context.Context, tenantID string) (string, error) {
	if s.pg == nil {
		return "", fmt.Errorf("postgres is not available")
	}
	var ciphertext, nonce []byte
	schema := quoteIdent(s.cfg.PostgresSchema)
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT gemini_key_ciphertext, gemini_key_nonce FROM %s.tenant_ai_configs WHERE tenant_id=$1`, schema), tenantID).Scan(&ciphertext, &nonce)
	if errors.Is(err, pgx.ErrNoRows) || len(ciphertext) == 0 {
		return "", nil
	}
	key, err := secretbox.ParseKey(s.cfg.TenantSecretEncryptionKey)
	if err != nil {
		return "", ErrTenantSecretNotConfigured
	}
	plaintext, err := secretbox.Decrypt(key, ciphertext, nonce)
	if err != nil {
		return "", ErrTenantSecretInvalid
	}
	return string(plaintext), nil
}

func (s *Store) GetTenantAgentConfig(ctx context.Context, tenantID, agentID string) (TenantAgentConfig, error) {
	if s.pg == nil {
		return TenantAgentConfig{}, fmt.Errorf("postgres is not available")
	}
	var out TenantAgentConfig
	schema := quoteIdent(s.cfg.PostgresSchema)
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT tenant_id, agent_id, system_prompt, enabled, updated_at FROM %s.tenant_agent_configs WHERE tenant_id=$1 AND agent_id=$2`, schema), tenantID, agentID).Scan(&out.TenantID, &out.AgentID, &out.SystemPrompt, &out.Enabled, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantAgentConfig{TenantID: tenantID, AgentID: agentID, Enabled: true}, nil
	}
	return out, err
}

func (s *Store) PutTenantAgentConfig(ctx context.Context, tenantID, agentID, prompt string, enabled bool) (TenantAgentConfig, error) {
	if len([]rune(prompt)) > 8000 {
		return TenantAgentConfig{}, ErrTenantPromptInvalid
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.tenant_agent_configs (tenant_id, agent_id, system_prompt, enabled, created_by, updated_by) VALUES ($1,$2,$3,$4,$5,$5) ON CONFLICT (tenant_id,agent_id) DO UPDATE SET system_prompt=EXCLUDED.system_prompt, enabled=EXCLUDED.enabled, updated_by=EXCLUDED.updated_by, updated_at=now()`, schema), tenantID, agentID, strings.TrimSpace(prompt), enabled, actor)
	if err != nil {
		return TenantAgentConfig{}, err
	}
	return s.GetTenantAgentConfig(ctx, tenantID, agentID)
}

func (s *Store) ListTenantTools(ctx context.Context, tenantID string) ([]TenantCallTool, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT id, tenant_id, tool_key, display_name, description, handler_key, input_schema, enabled, updated_at FROM %s.tenant_call_tools WHERE tenant_id=$1 ORDER BY display_name, tool_key`, schema), tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TenantCallTool
	for rows.Next() {
		var item TenantCallTool
		var raw []byte
		if err := rows.Scan(&item.ID, &item.TenantID, &item.ToolKey, &item.DisplayName, &item.Description, &item.HandlerKey, &raw, &item.Enabled, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &item.InputSchema); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListTenantAgentTools resolves only enabled tools assigned through enabled
// skills for the requested tenant and agent. The join predicates intentionally
// repeat tenant_id at every edge to prevent cross-tenant references.
func (s *Store) ListTenantAgentTools(ctx context.Context, tenantID, agentID string) ([]TenantCallTool, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT DISTINCT ON (t.id) t.id, t.tenant_id, t.tool_key, t.display_name, t.description, t.handler_key, t.input_schema, t.enabled, t.updated_at
FROM %s.tenant_call_tools t
JOIN %s.tenant_skill_tools st ON st.tenant_id=t.tenant_id AND st.tool_id=t.id
JOIN %s.tenant_skills sk ON sk.tenant_id=st.tenant_id AND sk.id=st.skill_id AND sk.enabled=true
JOIN %s.tenant_agent_skills ag ON ag.tenant_id=sk.tenant_id AND ag.skill_id=sk.id AND ag.agent_id=$2
WHERE t.tenant_id=$1 AND t.enabled=true
ORDER BY t.id`, schema, schema, schema, schema), tenantID, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TenantCallTool
	for rows.Next() {
		var item TenantCallTool
		var raw []byte
		if err := rows.Scan(&item.ID, &item.TenantID, &item.ToolKey, &item.DisplayName, &item.Description, &item.HandlerKey, &raw, &item.Enabled, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &item.InputSchema); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) CreateTenantTool(ctx context.Context, item TenantCallTool) (TenantCallTool, error) {
	if strings.TrimSpace(item.ToolKey) == "" || strings.TrimSpace(item.HandlerKey) == "" || len(item.InputSchema) == 0 {
		return TenantCallTool{}, ErrTenantToolInvalid
	}
	raw, err := json.Marshal(item.InputSchema)
	if err != nil || len(raw) > 32*1024 {
		return TenantCallTool{}, ErrTenantToolInvalid
	}
	actor := auditctx.ActorID(ctx)
	item.ID = newTenantAIID("tool")
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.tenant_call_tools (id, tenant_id, tool_key, display_name, description, handler_key, input_schema, enabled, created_by, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,$9,$9)`, schema), item.ID, item.TenantID, strings.TrimSpace(item.ToolKey), strings.TrimSpace(item.DisplayName), strings.TrimSpace(item.Description), strings.TrimSpace(item.HandlerKey), raw, item.Enabled, actor)
	if err != nil {
		return TenantCallTool{}, err
	}
	return s.getTenantTool(ctx, item.TenantID, item.ID)
}

func (s *Store) UpdateTenantTool(ctx context.Context, item TenantCallTool) (TenantCallTool, error) {
	raw, err := json.Marshal(item.InputSchema)
	if err != nil || len(raw) > 32*1024 || strings.TrimSpace(item.HandlerKey) == "" {
		return TenantCallTool{}, ErrTenantToolInvalid
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.tenant_call_tools SET tool_key=$3, display_name=$4, description=$5, handler_key=$6, input_schema=$7::jsonb, enabled=$8, updated_by=$9, updated_at=now() WHERE tenant_id=$1 AND id=$2`, schema), item.TenantID, item.ID, item.ToolKey, item.DisplayName, item.Description, item.HandlerKey, raw, item.Enabled, actor)
	if err != nil {
		return TenantCallTool{}, err
	}
	if tag.RowsAffected() == 0 {
		return TenantCallTool{}, ErrTenantAIConfigNotFound
	}
	return s.getTenantTool(ctx, item.TenantID, item.ID)
}

func (s *Store) DeleteTenantTool(ctx context.Context, tenantID, id string) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var used bool
	if err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s.tenant_skill_tools WHERE tenant_id=$1 AND tool_id=$2)`, schema), tenantID, id).Scan(&used); err != nil {
		return err
	}
	if used {
		return ErrTenantToolInUse
	}
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.tenant_call_tools WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTenantAIConfigNotFound
	}
	return nil
}

func (s *Store) getTenantTool(ctx context.Context, tenantID, id string) (TenantCallTool, error) {
	var item TenantCallTool
	var raw []byte
	schema := quoteIdent(s.cfg.PostgresSchema)
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id, tenant_id, tool_key, display_name, description, handler_key, input_schema, enabled, updated_at FROM %s.tenant_call_tools WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id).Scan(&item.ID, &item.TenantID, &item.ToolKey, &item.DisplayName, &item.Description, &item.HandlerKey, &raw, &item.Enabled, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantCallTool{}, ErrTenantAIConfigNotFound
	}
	if err != nil {
		return TenantCallTool{}, err
	}
	err = json.Unmarshal(raw, &item.InputSchema)
	return item, err
}

func (s *Store) ListTenantSkills(ctx context.Context, tenantID string) ([]TenantSkill, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT id, tenant_id, slug, name, prompt, enabled, updated_at FROM %s.tenant_skills WHERE tenant_id=$1 ORDER BY name, slug`, schema), tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TenantSkill
	for rows.Next() {
		var item TenantSkill
		if err := rows.Scan(&item.ID, &item.TenantID, &item.Slug, &item.Name, &item.Prompt, &item.Enabled, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if err := s.loadSkillLinks(ctx, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) CreateTenantSkill(ctx context.Context, item TenantSkill) (TenantSkill, error) {
	if len([]rune(item.Prompt)) > 8000 || strings.TrimSpace(item.Slug) == "" || strings.TrimSpace(item.Name) == "" {
		return TenantSkill{}, ErrTenantSkillInvalid
	}
	item.ID = newTenantAIID("skill")
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return TenantSkill{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.tenant_skills (id, tenant_id, slug, name, prompt, enabled, created_by, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$7)`, schema), item.ID, item.TenantID, item.Slug, item.Name, item.Prompt, item.Enabled, actor); err != nil {
		return TenantSkill{}, err
	}
	if err = replaceSkillLinks(ctx, tx, schema, item.TenantID, item.ID, item.ToolIDs, item.AgentIDs); err != nil {
		return TenantSkill{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return TenantSkill{}, err
	}
	return s.getTenantSkill(ctx, item.TenantID, item.ID)
}

func (s *Store) UpdateTenantSkill(ctx context.Context, item TenantSkill) (TenantSkill, error) {
	if len([]rune(item.Prompt)) > 8000 || strings.TrimSpace(item.Slug) == "" || strings.TrimSpace(item.Name) == "" {
		return TenantSkill{}, ErrTenantSkillInvalid
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return TenantSkill{}, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.tenant_skills SET slug=$3, name=$4, prompt=$5, enabled=$6, updated_by=$7, updated_at=now() WHERE tenant_id=$1 AND id=$2`, schema), item.TenantID, item.ID, item.Slug, item.Name, item.Prompt, item.Enabled, actor)
	if err != nil || tag.RowsAffected() == 0 {
		if err != nil {
			return TenantSkill{}, err
		}
		return TenantSkill{}, ErrTenantAIConfigNotFound
	}
	if err := replaceSkillLinks(ctx, tx, schema, item.TenantID, item.ID, item.ToolIDs, item.AgentIDs); err != nil {
		return TenantSkill{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return TenantSkill{}, err
	}
	return s.getTenantSkill(ctx, item.TenantID, item.ID)
}

func (s *Store) DeleteTenantSkill(ctx context.Context, tenantID, id string) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.tenant_skills WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTenantAIConfigNotFound
	}
	return nil
}

func (s *Store) UpdateTenantSkillAssignments(ctx context.Context, tenantID, skillID string, toolIDs, agentIDs []string) (TenantSkill, error) {
	if s.pg == nil {
		return TenantSkill{}, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return TenantSkill{}, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s.tenant_skills WHERE tenant_id=$1 AND id=$2)`, schema), tenantID, skillID).Scan(&exists); err != nil {
		return TenantSkill{}, err
	}
	if !exists {
		return TenantSkill{}, ErrTenantAIConfigNotFound
	}
	if err := replaceSkillLinks(ctx, tx, schema, tenantID, skillID, toolIDs, agentIDs); err != nil {
		return TenantSkill{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return TenantSkill{}, err
	}
	return s.getTenantSkill(ctx, tenantID, skillID)
}

func (s *Store) getTenantSkill(ctx context.Context, tenantID, id string) (TenantSkill, error) {
	var item TenantSkill
	schema := quoteIdent(s.cfg.PostgresSchema)
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id, tenant_id, slug, name, prompt, enabled, updated_at FROM %s.tenant_skills WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id).Scan(&item.ID, &item.TenantID, &item.Slug, &item.Name, &item.Prompt, &item.Enabled, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantSkill{}, ErrTenantAIConfigNotFound
	}
	if err != nil {
		return TenantSkill{}, err
	}
	if err := s.loadSkillLinks(ctx, &item); err != nil {
		return TenantSkill{}, err
	}
	return item, nil
}

func (s *Store) loadSkillLinks(ctx context.Context, item *TenantSkill) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT tool_id FROM %s.tenant_skill_tools WHERE tenant_id=$1 AND skill_id=$2 ORDER BY tool_id`, schema), item.TenantID, item.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		item.ToolIDs = append(item.ToolIDs, id)
	}
	rows.Close()
	rows, err = s.pg.Query(ctx, fmt.Sprintf(`SELECT agent_id FROM %s.tenant_agent_skills WHERE tenant_id=$1 AND skill_id=$2 ORDER BY agent_id`, schema), item.TenantID, item.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		item.AgentIDs = append(item.AgentIDs, id)
	}
	return rows.Err()
}

func replaceSkillLinks(ctx context.Context, tx pgx.Tx, schema, tenantID, skillID string, toolIDs, agentIDs []string) error {
	if len(toolIDs) > 20 || len(agentIDs) > 50 {
		return ErrTenantSkillInvalid
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.tenant_skill_tools WHERE tenant_id=$1 AND skill_id=$2`, schema), tenantID, skillID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s.tenant_agent_skills WHERE tenant_id=$1 AND skill_id=$2`, schema), tenantID, skillID); err != nil {
		return err
	}
	for _, toolID := range uniqueStrings(toolIDs) {
		var exists bool
		if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s.tenant_call_tools WHERE tenant_id=$1 AND id=$2 AND enabled=true)`, schema), tenantID, toolID).Scan(&exists); err != nil || !exists {
			if err != nil {
				return err
			}
			return ErrTenantSkillInvalid
		}
		if _, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.tenant_skill_tools (tenant_id, skill_id, tool_id, created_by, updated_by) VALUES ($1,$2,$3,'system','system')`, schema), tenantID, skillID, toolID); err != nil {
			return err
		}
	}
	for _, agentID := range uniqueStrings(agentIDs) {
		if _, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.tenant_agent_skills (tenant_id, agent_id, skill_id, created_by, updated_by) VALUES ($1,$2,$3,'system','system')`, schema), tenantID, agentID, skillID); err != nil {
			return err
		}
	}
	return nil
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func newTenantAIID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
