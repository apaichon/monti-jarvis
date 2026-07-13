package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/natsbus"
	"github.com/libra/monti-jarvis/internal/store"
)

type UserProfile struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        Role   `json:"role"`
	TenantID    string `json:"tenant_id,omitempty"`
}

type TokenPair struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	TokenType    string      `json:"token_type"`
	User         UserProfile `json:"user"`
}

type Dependencies struct {
	Store *store.Store
	Bus   *natsbus.Bus
	CH    *clickhouse.Client
	Cfg   Config
}

type Service struct {
	store        *store.Store
	cache        *Cache
	events       *EventPublisher
	persist      *PersistWorker
	issuer       *TokenIssuer
	refreshTTL   time.Duration
	accessTTL    time.Duration
	authDisabled bool
	writeBehind  bool
}

func NewService(deps Dependencies) (*Service, error) {
	cfg := deps.Cfg
	var issuer *TokenIssuer
	var err error
	if cfg.JWTSecret != "" {
		issuer, err = NewTokenIssuer(cfg.JWTSecret, cfg.AccessTTL)
		if err != nil {
			return nil, err
		}
	}
	cache := NewCache(deps.Store.Redis(), cfg.RedisPrefix, cfg.UserCacheTTL, cfg.RefreshTTL, cfg.CacheEnabled)
	eventsOn := cfg.EventsEnabled && deps.Bus != nil && deps.Bus.Enabled()
	svc := &Service{
		store:        deps.Store,
		cache:        cache,
		events:       NewEventPublisher(deps.Bus, deps.CH, eventsOn || (deps.CH != nil && deps.CH.Enabled())),
		persist:      NewPersistWorker(cache, deps.Store, cfg.WriteBehindEnabled),
		issuer:       issuer,
		refreshTTL:   cfg.RefreshTTL,
		accessTTL:    cfg.AccessTTL,
		authDisabled: cfg.AuthDisabled,
		writeBehind:  cfg.WriteBehindEnabled && cache.Enabled(),
	}
	return svc, nil
}

func (s *Service) Enabled() bool {
	return s != nil && !s.authDisabled && s.issuer != nil
}

func (s *Service) TokensEnabled() bool {
	return s != nil && s.issuer != nil
}

func (s *Service) Configured() bool {
	return s.TokensEnabled()
}

func (s *Service) Start(ctx context.Context) {
	if s.persist != nil {
		s.persist.Start(ctx)
	}
}

func (s *Service) CacheStatus() string {
	if s.cache != nil && s.cache.Enabled() {
		return "ok"
	}
	return "disabled"
}

func (s *Service) EventsStatus() string {
	if s.events != nil && s.events.Enabled() {
		return "ok"
	}
	if s.events != nil && s.events.ch != nil && s.events.ch.Enabled() {
		return "ok"
	}
	return "disabled"
}

func (s *Service) WriteBehindLag(ctx context.Context) int64 {
	if s.persist == nil {
		return 0
	}
	lag, _ := s.persist.Lag(ctx)
	return lag
}

func (s *Service) Login(ctx context.Context, email, password string) (TokenPair, error) {
	if !s.TokensEnabled() {
		return TokenPair{}, ErrNotConfigured
	}
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return TokenPair{}, ErrInvalidCredentials
	}
	meta := RequestMetaFrom(ctx)

	user, err := s.loadUserForLogin(ctx, email, true)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			s.events.PublishFailure(ctx, email, meta)
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, err
	}
	defer func() { _ = s.cache.StripPasswordHash(ctx, user) }()

	hash := user.PasswordHash
	if hash == "" {
		pg, err := s.store.GetUserByEmail(ctx, email)
		if err != nil {
			s.events.PublishFailure(ctx, email, meta)
			return TokenPair{}, ErrInvalidCredentials
		}
		hash = pg.PasswordHash
	}
	if user.Status != "active" || !VerifyPassword(hash, password) {
		s.events.PublishFailure(ctx, email, meta)
		return TokenPair{}, ErrInvalidCredentials
	}
	if !userEmailVerified(user) {
		s.events.PublishFailure(ctx, email, meta)
		return TokenPair{}, ErrEmailNotVerified
	}
	pair, err := s.issueForUser(ctx, user)
	if err != nil {
		return TokenPair{}, err
	}
	s.events.Publish(ctx, "auth.user.logged_in", user, meta, nil)
	return pair, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefresh string) (TokenPair, error) {
	return s.refresh(ctx, rawRefresh, "")
}

func (s *Service) RefreshWithAccess(ctx context.Context, rawRefresh, accessHeader string) (TokenPair, error) {
	return s.refresh(ctx, rawRefresh, accessHeader)
}

func (s *Service) refresh(ctx context.Context, rawRefresh, accessHeader string) (TokenPair, error) {
	if !s.TokensEnabled() {
		return TokenPair{}, ErrNotConfigured
	}
	if err := ValidateRefreshToken(rawRefresh); err != nil {
		return TokenPair{}, ErrUnauthorized
	}
	hash := HashRefreshToken(rawRefresh)
	row, err := s.loadRefresh(ctx, hash)
	if err != nil {
		return TokenPair{}, ErrUnauthorized
	}
	if row.RevokedAt != nil || time.Now().UTC().After(row.ExpiresAt) {
		return TokenPair{}, ErrUnauthorized
	}
	user, err := s.loadUserByID(ctx, row.UserID)
	if err != nil {
		return TokenPair{}, ErrUnauthorized
	}
	if user.Status != "active" {
		return TokenPair{}, ErrUnauthorized
	}
	_ = s.persistRefreshRevoke(ctx, hash)
	s.denyAccessHeader(ctx, accessHeader)
	s.events.Publish(ctx, "auth.token.revoked", user, RequestMetaFrom(ctx), map[string]any{"token_hash": hash})

	pair, err := s.issueForUser(ctx, user)
	if err != nil {
		return TokenPair{}, err
	}
	s.events.Publish(ctx, "auth.token.refreshed", user, RequestMetaFrom(ctx), nil)
	return pair, nil
}

func (s *Service) Logout(ctx context.Context, rawRefresh, accessHeader string) error {
	if !s.TokensEnabled() {
		return ErrNotConfigured
	}
	meta := RequestMetaFrom(ctx)
	if err := ValidateRefreshToken(rawRefresh); err == nil {
		hash := HashRefreshToken(rawRefresh)
		_ = s.persistRefreshRevoke(ctx, hash)
		if row, err := s.loadRefresh(ctx, hash); err == nil {
			if user, uerr := s.loadUserByID(ctx, row.UserID); uerr == nil {
				s.events.Publish(ctx, "auth.token.revoked", user, meta, map[string]any{"token_hash": hash})
			}
		}
	}
	s.denyAccessHeader(ctx, accessHeader)
	if accessHeader != "" {
		if ac, err := s.ParseBearer(accessHeader); err == nil {
			if user, uerr := s.loadUserByID(ctx, ac.UserID); uerr == nil {
				s.events.Publish(ctx, "auth.user.logged_out", user, meta, nil)
			}
		}
	}
	return nil
}

func (s *Service) IssueTokenPairForUser(ctx context.Context, user store.AuthUser) (TokenPair, error) {
	if !s.TokensEnabled() {
		return TokenPair{}, ErrNotConfigured
	}
	return s.issueForUser(ctx, cachedFromStore(user, false))
}

func (s *Service) IssueAccessForPrincipal(userID, email string, role Role, tenantID string) (token string, expiresIn int, err error) {
	if !s.TokensEnabled() {
		return "", 0, ErrNotConfigured
	}
	if !role.Valid() {
		return "", 0, fmt.Errorf("invalid role")
	}
	token, _, expiresIn, err = s.issuer.IssueAccess(userID, email, role, tenantID)
	return token, expiresIn, err
}

func (s *Service) RefreshTTLSeconds() int {
	if s == nil || s.refreshTTL <= 0 {
		return 0
	}
	return int(s.refreshTTL.Seconds())
}

func (s *Service) Me(ctx context.Context, userID string) (UserProfile, error) {
	if !s.TokensEnabled() {
		return UserProfile{}, ErrNotConfigured
	}
	user, err := s.loadUserByID(ctx, userID)
	if err != nil {
		return UserProfile{}, ErrUnauthorized
	}
	return profileFromCached(user), nil
}

func (s *Service) ParseBearer(authHeader string) (AuthContext, error) {
	if !s.TokensEnabled() {
		return AuthContext{}, ErrNotConfigured
	}
	token, ok := bearerToken(authHeader)
	if !ok {
		return AuthContext{}, ErrUnauthorized
	}
	ac, exp, err := s.issuer.ParseAccess(token)
	if err != nil {
		return AuthContext{}, ErrUnauthorized
	}
	if s.cache != nil && s.cache.Enabled() && ac.JTI != "" {
		denied, derr := s.cache.IsJTIDenied(context.Background(), ac.JTI)
		if derr != nil {
			return AuthContext{}, derr
		}
		if denied {
			return AuthContext{}, ErrUnauthorized
		}
	}
	if !exp.IsZero() && time.Now().UTC().After(exp) {
		return AuthContext{}, ErrUnauthorized
	}
	return ac, nil
}

func (s *Service) denyAccessHeader(ctx context.Context, authHeader string) {
	token, ok := bearerToken(authHeader)
	if !ok {
		return
	}
	ac, exp, err := s.issuer.ParseAccess(token)
	if err != nil || ac.JTI == "" || s.cache == nil || !s.cache.Enabled() {
		return
	}
	ttl := time.Until(exp)
	if ttl <= 0 {
		ttl = s.accessTTL
	}
	_ = s.cache.DenyJTI(ctx, ac.JTI, ttl)
}

func (s *Service) loadUserForLogin(ctx context.Context, email string, includeHash bool) (CachedUser, error) {
	if s.cache != nil && s.cache.Enabled() {
		if id, ok, err := s.cache.GetUserIDByEmail(ctx, email); err != nil {
			return CachedUser{}, err
		} else if ok {
			if user, hit, err := s.cache.GetUserByID(ctx, id); err != nil {
				return CachedUser{}, err
			} else if hit {
				return user, nil
			}
		}
	}
	pg, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return CachedUser{}, err
	}
	user := cachedFromStore(pg, includeHash)
	_ = s.cache.PutUser(ctx, user, includeHash)
	return user, nil
}

func (s *Service) loadUserByID(ctx context.Context, userID string) (CachedUser, error) {
	if s.cache != nil && s.cache.Enabled() {
		if user, ok, err := s.cache.GetUserByID(ctx, userID); err != nil {
			return CachedUser{}, err
		} else if ok && user.Status != "" {
			return user, nil
		}
	}
	pg, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return CachedUser{}, err
	}
	user := cachedFromStore(pg, false)
	_ = s.cache.PutUser(ctx, user, false)
	return user, nil
}

func (s *Service) loadRefresh(ctx context.Context, hash string) (store.RefreshTokenRow, error) {
	if s.cache != nil && s.cache.Enabled() {
		if row, ok, err := s.cache.GetRefresh(ctx, hash); err != nil {
			return store.RefreshTokenRow{}, err
		} else if ok {
			if row.Revoked {
				return store.RefreshTokenRow{RevokedAt: ptrTime(time.Now())}, nil
			}
			return store.RefreshTokenRow{
				ID:        row.ID,
				UserID:    row.UserID,
				TokenHash: hash,
				ExpiresAt: row.ExpiresAt,
			}, nil
		}
	}
	pg, err := s.store.GetRefreshToken(ctx, hash)
	if err != nil {
		return store.RefreshTokenRow{}, err
	}
	_ = s.cache.PutRefresh(ctx, hash, CachedRefresh{
		ID:        pg.ID,
		UserID:    pg.UserID,
		ExpiresAt: pg.ExpiresAt,
		Revoked:   pg.RevokedAt != nil,
	})
	return pg, nil
}

func (s *Service) issueForUser(ctx context.Context, user CachedUser) (TokenPair, error) {
	role := Role(user.Role)
	if !role.Valid() {
		return TokenPair{}, fmt.Errorf("user has no valid role")
	}
	access, _, expiresIn, err := s.issuer.IssueAccess(user.ID, user.Email, role, user.TenantID)
	if err != nil {
		return TokenPair{}, err
	}
	raw, hash, err := NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}
	expiresAt := time.Now().UTC().Add(s.refreshTTL)
	tokenID := newID()
	if err := s.persistRefreshCreate(ctx, tokenID, user.ID, hash, expiresAt); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
		User:         profileFromCached(user),
	}, nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func bearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}
