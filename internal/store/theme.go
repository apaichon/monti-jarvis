package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/minio/minio-go/v7"
)

var (
	ErrThemeNotFound              = errors.New("theme not found")
	ErrInvalidThemeTokens         = errors.New("invalid_theme_tokens")
	ErrInvalidThemePreset         = errors.New("invalid_preset")
	ErrInvalidThemeBranding       = errors.New("invalid_theme_branding")
	ErrContrastConfirmationNeeded = errors.New("contrast_confirmation_required")
)

// Theme branding shown on caller desk / embed header.
type ThemeBranding struct {
	BrandName string `json:"brand_name"`
	Subtitle  string `json:"subtitle"`
	LogoURL   string `json:"logo_url"`
	LogoAlt   string `json:"logo_alt"`
}

// ThemeTokens maps semantic keys to #RRGGBB colors.
type ThemeTokens map[string]string

// TenantTheme is one row of callcenter.tenant_themes.
type TenantTheme struct {
	TenantID           string        `json:"tenant_id"`
	Preset             string        `json:"preset"`
	DraftBranding      ThemeBranding `json:"draft_branding"`
	PublishedBranding  ThemeBranding `json:"published_branding"`
	DraftTokens        ThemeTokens   `json:"draft_tokens"`
	PublishedTokens    ThemeTokens   `json:"published_tokens"`
	PublishedAt        *time.Time    `json:"published_at,omitempty"`
	DraftUpdatedAt     time.Time     `json:"draft_updated_at,omitempty"`
	CreatedAt          time.Time     `json:"created_at,omitempty"`
	UpdatedAt          time.Time     `json:"updated_at,omitempty"`
	ContrastReport     ContrastReport `json:"contrast_report,omitempty"`
}

// ContrastPair is one checked pair.
type ContrastPair struct {
	Pair  string  `json:"pair"`
	Ratio float64 `json:"ratio"`
	Pass  bool    `json:"pass"`
}

// ContrastReport summarizes WCAG-style checks on draft (or given) tokens.
type ContrastReport struct {
	OK    bool           `json:"ok"`
	Pairs []ContrastPair `json:"pairs"`
}

// PublicTheme is the published payload for clients.
type PublicTheme struct {
	TenantID  string        `json:"tenant_id"`
	Preset    string        `json:"preset"`
	Source    string        `json:"source"` // published | system_default
	Branding  ThemeBranding `json:"branding"`
	Tokens    ThemeTokens   `json:"tokens"`
}

var requiredThemeTokenKeys = []string{
	"primary", "primary_text", "accent", "background", "surface", "surface_elevated",
	"text", "muted", "line", "success", "warn", "danger",
}

var hexColorRE = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// DefaultDarkTokens matches pre-S39 Monti dark palette (approx).
func DefaultDarkTokens() ThemeTokens {
	return ThemeTokens{
		"primary":          "#006dff",
		"primary_text":     "#ffffff",
		"accent":           "#00b7ff",
		"background":       "#020712",
		"surface":          "#05101f",
		"surface_elevated": "#08172a",
		"text":             "#f7fbff",
		"muted":            "#8fa5bf",
		"line":             "#3d5a80",
		"success":          "#12f2a5",
		"warn":             "#f0b83f",
		"danger":           "#ff3b3b",
	}
}

// DefaultLightTokens is the light preset.
func DefaultLightTokens() ThemeTokens {
	return ThemeTokens{
		"primary":          "#006dff",
		"primary_text":     "#ffffff",
		"accent":           "#0084ff",
		"background":       "#f4f7fb",
		"surface":          "#ffffff",
		"surface_elevated": "#eef3f9",
		"text":             "#0b1220",
		"muted":            "#5a6b82",
		"line":             "#c5d0e0",
		"success":          "#0f9f6e",
		"warn":             "#c98a10",
		"danger":           "#d92d20",
	}
}

// DefaultBrandingFallback returns empty branding (client applies defaults).
func DefaultBrandingFallback() ThemeBranding {
	return ThemeBranding{}
}

func PresetTokens(preset string) (ThemeTokens, error) {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "", "dark":
		return DefaultDarkTokens(), nil
	case "light":
		return DefaultLightTokens(), nil
	case "branded":
		// branded starts from dark base
		return DefaultDarkTokens(), nil
	default:
		return nil, ErrInvalidThemePreset
	}
}

func NormalizePreset(preset string) (string, error) {
	p := strings.ToLower(strings.TrimSpace(preset))
	if p == "" {
		p = "dark"
	}
	switch p {
	case "dark", "light", "branded":
		return p, nil
	default:
		return "", ErrInvalidThemePreset
	}
}

// ExpandHexColor normalizes #RGB / #RRGGBB to lowercase #RRGGBB.
func ExpandHexColor(v string) (string, error) {
	v = strings.TrimSpace(v)
	if !hexColorRE.MatchString(v) {
		return "", fmt.Errorf("%w: %s", ErrInvalidThemeTokens, v)
	}
	h := strings.TrimPrefix(v, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	return "#" + strings.ToLower(h), nil
}

// ValidateAndNormalizeTokens ensures all required keys and hex colors.
func ValidateAndNormalizeTokens(in ThemeTokens) (ThemeTokens, error) {
	if in == nil {
		return nil, fmt.Errorf("%w: empty", ErrInvalidThemeTokens)
	}
	out := ThemeTokens{}
	for _, k := range requiredThemeTokenKeys {
		raw, ok := in[k]
		if !ok || strings.TrimSpace(raw) == "" {
			return nil, fmt.Errorf("%w: missing %s", ErrInvalidThemeTokens, k)
		}
		hex, err := ExpandHexColor(raw)
		if err != nil {
			return nil, fmt.Errorf("%w: bad %s", ErrInvalidThemeTokens, k)
		}
		out[k] = hex
	}
	// reject unknown keys? allow extras ignored for forward compat
	return out, nil
}

// ValidateAndNormalizeBranding cleans branding fields.
func ValidateAndNormalizeBranding(in ThemeBranding) (ThemeBranding, error) {
	out := ThemeBranding{
		BrandName: strings.TrimSpace(in.BrandName),
		Subtitle:  strings.TrimSpace(in.Subtitle),
		LogoURL:   strings.TrimSpace(in.LogoURL),
		LogoAlt:   strings.TrimSpace(in.LogoAlt),
	}
	if len(out.BrandName) > 80 {
		return out, fmt.Errorf("%w: brand_name too long", ErrInvalidThemeBranding)
	}
	if len(out.Subtitle) > 120 {
		return out, fmt.Errorf("%w: subtitle too long", ErrInvalidThemeBranding)
	}
	if len(out.LogoAlt) > 80 {
		return out, fmt.Errorf("%w: logo_alt too long", ErrInvalidThemeBranding)
	}
	if out.LogoURL != "" {
		if len(out.LogoURL) > 2048 {
			return out, fmt.Errorf("%w: logo_url too long", ErrInvalidThemeBranding)
		}
		if strings.HasPrefix(out.LogoURL, "/") {
			// same-origin path ok
		} else {
			u, err := url.Parse(out.LogoURL)
			if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
				return out, fmt.Errorf("%w: logo_url must be http(s) or path", ErrInvalidThemeBranding)
			}
		}
	}
	return out, nil
}

// ContrastRatio returns WCAG contrast ratio of two #RRGGBB colors.
func ContrastRatio(hexA, hexB string) (float64, error) {
	a, err := ExpandHexColor(hexA)
	if err != nil {
		return 0, err
	}
	b, err := ExpandHexColor(hexB)
	if err != nil {
		return 0, err
	}
	la := relativeLuminance(a)
	lb := relativeLuminance(b)
	lighter := math.Max(la, lb)
	darker := math.Min(la, lb)
	return (lighter + 0.05) / (darker + 0.05), nil
}

func relativeLuminance(hex string) float64 {
	h := strings.TrimPrefix(hex, "#")
	r := hexByte(h[0:2])
	g := hexByte(h[2:4])
	b := hexByte(h[4:6])
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

func hexByte(s string) float64 {
	var v int
	fmt.Sscanf(s, "%x", &v)
	return float64(v) / 255.0
}

func lin(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// EvaluateContrast checks required pairs on tokens.
func EvaluateContrast(tokens ThemeTokens) ContrastReport {
	report := ContrastReport{OK: true, Pairs: []ContrastPair{}}
	check := func(name, fg, bg string, min float64) {
		ratio, err := ContrastRatio(tokens[fg], tokens[bg])
		if err != nil {
			report.OK = false
			report.Pairs = append(report.Pairs, ContrastPair{Pair: name, Ratio: 0, Pass: false})
			return
		}
		pass := ratio >= min
		if !pass {
			report.OK = false
		}
		report.Pairs = append(report.Pairs, ContrastPair{Pair: name, Ratio: math.Round(ratio*100) / 100, Pass: pass})
	}
	check("text_on_surface", "text", "surface", 4.5)
	check("text_on_background", "text", "background", 4.5)
	check("primary_text_on_primary", "primary_text", "primary", 4.5)
	check("muted_on_surface", "muted", "surface", 3.0)
	return report
}

// ResolvePublicBranding applies fallbacks for empty published fields.
func ResolvePublicBranding(b ThemeBranding, workspaceName string) ThemeBranding {
	out := b
	if strings.TrimSpace(out.BrandName) == "" {
		out.BrandName = strings.TrimSpace(workspaceName)
		if out.BrandName == "" {
			out.BrandName = "Monti"
		}
	}
	if strings.TrimSpace(out.Subtitle) == "" {
		out.Subtitle = "AI · text & voice"
	}
	if strings.TrimSpace(out.LogoURL) == "" {
		out.LogoURL = "/images/monti-logo.png"
	}
	if strings.TrimSpace(out.LogoAlt) == "" {
		out.LogoAlt = out.BrandName
	}
	return out
}

// CSSVarMap returns CSS custom property map for clients.
func CSSVarMap(tokens ThemeTokens) map[string]string {
	m := map[string]string{}
	for k, v := range tokens {
		m["--mj-"+strings.ReplaceAll(k, "_", "-")] = v
	}
	// bridge to legacy customer-web vars when applied
	if v, ok := tokens["background"]; ok {
		m["--bg"] = v
	}
	if v, ok := tokens["text"]; ok {
		m["--ink"] = v
	}
	if v, ok := tokens["muted"]; ok {
		m["--muted"] = v
	}
	if v, ok := tokens["surface"]; ok {
		m["--panel"] = v
	}
	if v, ok := tokens["surface_elevated"]; ok {
		m["--panel-2"] = v
	}
	if v, ok := tokens["line"]; ok {
		m["--line"] = v
	}
	if v, ok := tokens["accent"]; ok {
		m["--cyan"] = v
	}
	if v, ok := tokens["primary"]; ok {
		m["--blue"] = v
	}
	if v, ok := tokens["success"]; ok {
		m["--green"] = v
	}
	if v, ok := tokens["danger"]; ok {
		m["--red"] = v
	}
	return m
}

func (s *Store) ensureThemeSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_themes (
  tenant_id text PRIMARY KEY REFERENCES %s.tenants(id) ON DELETE CASCADE,
  preset text NOT NULL DEFAULT 'dark',
  draft_branding jsonb NOT NULL DEFAULT '{}'::jsonb,
  published_branding jsonb NOT NULL DEFAULT '{}'::jsonb,
  draft_tokens jsonb NOT NULL DEFAULT '{}'::jsonb,
  published_tokens jsonb NOT NULL DEFAULT '{}'::jsonb,
  published_at timestamptz,
  draft_updated_at timestamptz NOT NULL DEFAULT now(),%s
)`, schema, schema, auditColumnsDDL),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("theme schema: %w", err)
		}
	}
	return nil
}

func emptyBrandingJSON() []byte { return []byte(`{}`) }

func brandingToJSON(b ThemeBranding) ([]byte, error) {
	return json.Marshal(b)
}

func tokensToJSON(t ThemeTokens) ([]byte, error) {
	if t == nil {
		t = ThemeTokens{}
	}
	return json.Marshal(t)
}

func parseBranding(raw []byte) ThemeBranding {
	var b ThemeBranding
	if len(raw) == 0 {
		return b
	}
	_ = json.Unmarshal(raw, &b)
	return b
}

func parseTokens(raw []byte) ThemeTokens {
	var t ThemeTokens
	if len(raw) == 0 {
		return ThemeTokens{}
	}
	_ = json.Unmarshal(raw, &t)
	if t == nil {
		return ThemeTokens{}
	}
	return t
}

func (s *Store) scanTheme(ctx context.Context, q string, arg any) (*TenantTheme, error) {
	var row TenantTheme
	var draftB, pubB, draftT, pubT []byte
	var publishedAt *time.Time
	err := s.pg.QueryRow(ctx, q, arg).Scan(
		&row.TenantID, &row.Preset,
		&draftB, &pubB, &draftT, &pubT,
		&publishedAt, &row.DraftUpdatedAt, &row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrThemeNotFound
	}
	if err != nil {
		return nil, err
	}
	row.DraftBranding = parseBranding(draftB)
	row.PublishedBranding = parseBranding(pubB)
	row.DraftTokens = parseTokens(draftT)
	row.PublishedTokens = parseTokens(pubT)
	row.PublishedAt = publishedAt
	if len(row.DraftTokens) == 0 {
		row.DraftTokens = DefaultDarkTokens()
	}
	row.ContrastReport = EvaluateContrast(row.DraftTokens)
	return &row, nil
}

// GetTenantTheme returns theme or ErrThemeNotFound.
func (s *Store) GetTenantTheme(ctx context.Context, tenantID string) (*TenantTheme, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return s.scanTheme(ctx, fmt.Sprintf(`
SELECT tenant_id, preset, draft_branding, published_branding, draft_tokens, published_tokens,
       published_at, draft_updated_at, created_at, updated_at
FROM %s.tenant_themes WHERE tenant_id = $1`, schema), tenantID)
}

// GetOrCreateTenantTheme lazy-creates dark defaults.
func (s *Store) GetOrCreateTenantTheme(ctx context.Context, tenantID string) (*TenantTheme, error) {
	row, err := s.GetTenantTheme(ctx, tenantID)
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, ErrThemeNotFound) {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tokens, _ := tokensToJSON(DefaultDarkTokens())
	brand, _ := brandingToJSON(ThemeBranding{})
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_themes (
  tenant_id, preset, draft_branding, published_branding, draft_tokens, published_tokens, created_by, updated_by
) VALUES ($1, 'dark', $2, $3, $4, $5, $6, $6)
ON CONFLICT (tenant_id) DO NOTHING`, schema),
		tenantID, brand, emptyBrandingJSON(), tokens, []byte(`{}`), actor)
	if err != nil {
		return nil, err
	}
	return s.GetTenantTheme(ctx, tenantID)
}

// UpdateTenantThemeDraft saves draft branding + tokens.
func (s *Store) UpdateTenantThemeDraft(ctx context.Context, tenantID string, preset string, branding ThemeBranding, tokens ThemeTokens) (*TenantTheme, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	if _, err := s.GetOrCreateTenantTheme(ctx, tenantID); err != nil {
		return nil, err
	}
	p, err := NormalizePreset(preset)
	if err != nil {
		return nil, err
	}
	b, err := ValidateAndNormalizeBranding(branding)
	if err != nil {
		return nil, err
	}
	// Allow partial token maps by merging over current draft.
	cur, err := s.GetTenantTheme(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	merged := ThemeTokens{}
	for k, v := range cur.DraftTokens {
		merged[k] = v
	}
	for k, v := range tokens {
		if strings.TrimSpace(v) == "" {
			continue
		}
		merged[k] = v
	}
	// If preset changed and tokens empty-ish custom, fill from preset base
	if p != cur.Preset && len(tokens) == 0 {
		base, _ := PresetTokens(p)
		merged = base
	}
	norm, err := ValidateAndNormalizeTokens(merged)
	if err != nil {
		return nil, err
	}
	bj, err := brandingToJSON(b)
	if err != nil {
		return nil, err
	}
	tj, err := tokensToJSON(norm)
	if err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_themes
SET preset = $2, draft_branding = $3, draft_tokens = $4,
    draft_updated_at = now(), updated_by = $5, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, p, bj, tj, actor)
	if err != nil {
		return nil, err
	}
	return s.GetTenantTheme(ctx, tenantID)
}

// PublishTenantTheme copies draft → published.
func (s *Store) PublishTenantTheme(ctx context.Context, tenantID string, confirmLowContrast bool) (*TenantTheme, error) {
	cur, err := s.GetOrCreateTenantTheme(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	tokens, err := ValidateAndNormalizeTokens(cur.DraftTokens)
	if err != nil {
		return nil, err
	}
	branding, err := ValidateAndNormalizeBranding(cur.DraftBranding)
	if err != nil {
		return nil, err
	}
	report := EvaluateContrast(tokens)
	if !report.OK && !confirmLowContrast {
		return nil, ErrContrastConfirmationNeeded
	}
	bj, _ := brandingToJSON(branding)
	tj, _ := tokensToJSON(tokens)
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_themes
SET published_branding = $2, published_tokens = $3, published_at = now(),
    draft_branding = $2, draft_tokens = $3,
    updated_by = $4, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, bj, tj, actor)
	if err != nil {
		return nil, err
	}
	return s.GetTenantTheme(ctx, tenantID)
}

// ResetTenantThemeDraft resets draft tokens to preset; optionally clears branding.
func (s *Store) ResetTenantThemeDraft(ctx context.Context, tenantID, preset string, resetBranding bool) (*TenantTheme, error) {
	if _, err := s.GetOrCreateTenantTheme(ctx, tenantID); err != nil {
		return nil, err
	}
	p, err := NormalizePreset(preset)
	if err != nil {
		return nil, err
	}
	base, err := PresetTokens(p)
	if err != nil {
		return nil, err
	}
	cur, err := s.GetTenantTheme(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	branding := cur.DraftBranding
	if resetBranding {
		branding = ThemeBranding{}
	}
	return s.UpdateTenantThemeDraft(ctx, tenantID, p, branding, base)
}

// GetPublishedTheme returns public payload with fallbacks.
func (s *Store) GetPublishedTheme(ctx context.Context, tenantID string) (*PublicTheme, error) {
	workspace := ""
	if t, err := s.GetTenant(ctx, tenantID); err == nil {
		workspace = t.Name
	}
	row, err := s.GetTenantTheme(ctx, tenantID)
	if err != nil {
		if errors.Is(err, ErrThemeNotFound) {
			return &PublicTheme{
				TenantID: tenantID,
				Preset:   "dark",
				Source:   "system_default",
				Branding: ResolvePublicBranding(ThemeBranding{}, workspace),
				Tokens:   DefaultDarkTokens(),
			}, nil
		}
		return nil, err
	}
	if len(row.PublishedTokens) == 0 {
		return &PublicTheme{
			TenantID: tenantID,
			Preset:   row.Preset,
			Source:   "system_default",
			Branding: ResolvePublicBranding(row.PublishedBranding, workspace),
			Tokens:   DefaultDarkTokens(),
		}, nil
	}
	tokens, err := ValidateAndNormalizeTokens(row.PublishedTokens)
	if err != nil {
		tokens = DefaultDarkTokens()
	}
	return &PublicTheme{
		TenantID: tenantID,
		Preset:   row.Preset,
		Source:   "published",
		Branding: ResolvePublicBranding(row.PublishedBranding, workspace),
		Tokens:   tokens,
	}, nil
}

// --- Logo assets ---

const themeAssetsPrefix = "theme/"

func ThemeLogoKey(tenantID, ext string) string {
	id := strings.TrimSpace(tenantID)
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	if ext == "" {
		ext = "png"
	}
	return themeAssetsPrefix + id + "/logo." + ext
}

func ThemeLogoURL(tenantID, ext string) string {
	id := strings.TrimSpace(tenantID)
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	if ext == "" {
		ext = "png"
	}
	return "/api/assets/theme/" + id + "/logo." + ext
}

// PutThemeLogo stores logo in MinIO and returns object key + public path URL.
func (s *Store) PutThemeLogo(ctx context.Context, tenantID, contentType string, data []byte) (string, string, error) {
	if s.minio == nil {
		return "", "", fmt.Errorf("minio is not available")
	}
	ext := contentTypeToExt(contentType)
	if ext == "" {
		return "", "", fmt.Errorf("unsupported image type")
	}
	key := ThemeLogoKey(tenantID, ext)
	_, err := s.minio.PutObject(ctx, s.cfg.MinioBucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", err
	}
	return key, ThemeLogoURL(tenantID, ext), nil
}

// GetThemeAsset streams a theme logo object.
func (s *Store) GetThemeAsset(ctx context.Context, tenantID, filename string) (io.ReadCloser, string, error) {
	if s.minio == nil {
		return nil, "", fmt.Errorf("minio is not available")
	}
	id := strings.TrimSpace(tenantID)
	name := path.Base(strings.TrimSpace(filename))
	if id == "" || name == "" || name == "." || strings.Contains(name, "..") {
		return nil, "", fmt.Errorf("invalid asset path")
	}
	if !strings.HasPrefix(name, "logo.") {
		return nil, "", fmt.Errorf("invalid asset filename")
	}
	key := themeAssetsPrefix + id + "/" + name
	obj, err := s.minio.GetObject(ctx, s.cfg.MinioBucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	stat, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, "", err
	}
	ct := stat.ContentType
	if ct == "" {
		ct = extToContentType(path.Ext(name))
	}
	return obj, ct, nil
}
