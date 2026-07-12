package tenantoauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var ErrSessionExpired = errors.New("oauth session expired")

type Config struct {
	PublicBaseURL      string
	GoogleClientID     string
	GoogleClientSecret string
	// GoogleRedirectURL optional full callback URI registered in Google Cloud Console.
	GoogleRedirectURL  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	RedisPrefix        string
}

type Service struct {
	cfg    Config
	redis  *redis.Client
	google *oauth2.Config
	github *oauth2.Config
}

type StartParams struct {
	Provider    string
	CompanyName string
	Slug        string
	DisplayName string
}

type StatePayload struct {
	Provider    string `json:"provider"`
	CompanyName string `json:"company_name,omitempty"`
	Slug        string `json:"slug,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type Identity struct {
	Provider       string
	ProviderUserID string
	Email          string
	DisplayName    string
}

type PendingSession struct {
	ID             string
	Provider       string
	ProviderUserID string
	Email          string
	DisplayName    string
}

func New(redis *redis.Client, cfg Config) *Service {
	if cfg.RedisPrefix == "" {
		cfg.RedisPrefix = "monti_jarvis:"
	}
	s := &Service{cfg: cfg, redis: redis}
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		redir := resolveOAuthRedirectURL(cfg.PublicBaseURL, cfg.GoogleRedirectURL, "google")
		log.Printf("oauth: google redirect_uri=%s", redir)
		s.google = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  redir,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}
	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" {
		redir := resolveOAuthRedirectURL(cfg.PublicBaseURL, cfg.GitHubRedirectURL, "github")
		log.Printf("oauth: github redirect_uri=%s", redir)
		s.github = &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  redir,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		}
	}
	return s
}

// OAuthCallbackPath is the shared Google/GitHub callback for login and register.
// Example: /api/public/tenant/oauth/google/callback
func OAuthCallbackPath(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	return "/api/public/tenant/oauth/" + provider + "/callback"
}

// resolveOAuthRedirectURL picks the callback URI for provider.
// Explicit override wins. Otherwise PublicBaseURL + path; for plain HTTP on a
// non-loopback host (e.g. http://monti-jarvis-dev.local:8091) Google returns
// invalid_request — rewrite to http://localhost:PORT/... which Google allows.
func resolveOAuthRedirectURL(publicBase, explicit, provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	path := OAuthCallbackPath(provider)
	if e := strings.TrimSpace(explicit); e != "" {
		return strings.TrimRight(e, "/")
	}
	base := strings.TrimRight(strings.TrimSpace(publicBase), "/")
	if base == "" {
		base = "http://localhost:8091"
	}
	if u, err := url.Parse(base); err == nil && u.Scheme != "" && u.Host != "" {
		host := strings.ToLower(u.Hostname())
		if strings.EqualFold(u.Scheme, "http") && host != "localhost" && host != "127.0.0.1" {
			port := u.Port()
			if port == "" {
				port = "80"
			}
			loop := "http://localhost"
			if port != "80" {
				loop = "http://localhost:" + port
			}
			log.Printf("oauth: %s using loopback redirect %s (APP_PUBLIC_URL host %q is not allowed by Google over http)", provider, loop+path, host)
			return loop + path
		}
	}
	return base + path
}

func (s *Service) ProviderEnabled(provider string) bool {
	switch strings.ToLower(provider) {
	case "google":
		return s.google != nil
	case "github":
		return s.github != nil
	default:
		return false
	}
}

func (s *Service) StartURL(ctx context.Context, params StartParams) (string, error) {
	cfg, err := s.config(params.Provider)
	if err != nil {
		return "", err
	}
	state, err := s.saveState(ctx, StatePayload{
		Provider:    strings.ToLower(params.Provider),
		CompanyName: strings.TrimSpace(params.CompanyName),
		Slug:        strings.TrimSpace(strings.ToLower(params.Slug)),
		DisplayName: strings.TrimSpace(params.DisplayName),
	})
	if err != nil {
		return "", err
	}
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

func (s *Service) Exchange(ctx context.Context, provider, state, code string) (Identity, StatePayload, error) {
	payload, err := s.consumeState(ctx, state)
	if err != nil {
		return Identity{}, StatePayload{}, err
	}
	if payload.Provider != strings.ToLower(provider) {
		return Identity{}, StatePayload{}, fmt.Errorf("oauth provider mismatch")
	}
	cfg, err := s.config(provider)
	if err != nil {
		return Identity{}, StatePayload{}, err
	}
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return Identity{}, StatePayload{}, err
	}
	identity, err := s.fetchIdentity(ctx, provider, token)
	if err != nil {
		return Identity{}, StatePayload{}, err
	}
	if payload.DisplayName != "" {
		identity.DisplayName = payload.DisplayName
	}
	return identity, payload, nil
}

func (s *Service) CreatePendingSession(ctx context.Context, identity Identity) (PendingSession, error) {
	id, err := newID()
	if err != nil {
		return PendingSession{}, err
	}
	session := PendingSession{
		ID:             id,
		Provider:       identity.Provider,
		ProviderUserID: identity.ProviderUserID,
		Email:          identity.Email,
		DisplayName:    identity.DisplayName,
	}
	if s.redis == nil {
		return PendingSession{}, fmt.Errorf("redis unavailable")
	}
	raw, _ := json.Marshal(session)
	key := s.cfg.RedisPrefix + "oauth_reg:" + id
	if err := s.redis.Set(ctx, key, raw, 15*time.Minute).Err(); err != nil {
		return PendingSession{}, err
	}
	return session, nil
}

func (s *Service) LoadPendingSession(ctx context.Context, sessionID string) (PendingSession, error) {
	if s.redis == nil {
		return PendingSession{}, fmt.Errorf("redis unavailable")
	}
	key := s.cfg.RedisPrefix + "oauth_reg:" + strings.TrimSpace(sessionID)
	raw, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return PendingSession{}, ErrSessionExpired
		}
		return PendingSession{}, err
	}
	var session PendingSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return PendingSession{}, err
	}
	return session, nil
}

func (s *Service) DeletePendingSession(ctx context.Context, sessionID string) {
	if s.redis == nil {
		return
	}
	_ = s.redis.Del(ctx, s.cfg.RedisPrefix+"oauth_reg:"+strings.TrimSpace(sessionID)).Err()
}

func (s *Service) config(provider string) (*oauth2.Config, error) {
	switch strings.ToLower(provider) {
	case "google":
		if s.google == nil {
			return nil, fmt.Errorf("google oauth is not configured")
		}
		return s.google, nil
	case "github":
		if s.github == nil {
			return nil, fmt.Errorf("github oauth is not configured")
		}
		return s.github, nil
	default:
		return nil, fmt.Errorf("unsupported oauth provider")
	}
}

func (s *Service) saveState(ctx context.Context, payload StatePayload) (string, error) {
	if s.redis == nil {
		return "", fmt.Errorf("redis unavailable")
	}
	state, err := newID()
	if err != nil {
		return "", err
	}
	raw, _ := json.Marshal(payload)
	key := s.cfg.RedisPrefix + "oauth_state:" + state
	if err := s.redis.Set(ctx, key, raw, 10*time.Minute).Err(); err != nil {
		return "", err
	}
	return state, nil
}

func (s *Service) consumeState(ctx context.Context, state string) (StatePayload, error) {
	if s.redis == nil {
		return StatePayload{}, fmt.Errorf("redis unavailable")
	}
	key := s.cfg.RedisPrefix + "oauth_state:" + strings.TrimSpace(state)
	raw, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return StatePayload{}, ErrSessionExpired
		}
		return StatePayload{}, err
	}
	_ = s.redis.Del(ctx, key).Err()
	var payload StatePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return StatePayload{}, err
	}
	return payload, nil
}

func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}