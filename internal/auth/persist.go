package auth

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/redis/go-redis/v9"
)

const persistGroup = "auth-persist"

type PersistWorker struct {
	cache   *Cache
	store   *store.Store
	enabled bool
}

func NewPersistWorker(cache *Cache, st *store.Store, enabled bool) *PersistWorker {
	return &PersistWorker{cache: cache, store: st, enabled: enabled && cache != nil && cache.Enabled()}
}

func (w *PersistWorker) Start(ctx context.Context) {
	if w == nil || !w.enabled {
		return
	}
	stream := w.cache.key(persistStream)
	_ = w.cache.rdb.XGroupCreateMkStream(ctx, stream, persistGroup, "0").Err()
	go w.loop(ctx, stream)
}

func (w *PersistWorker) loop(ctx context.Context, stream string) {
	consumer := "worker-1"
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		streams, err := w.cache.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    persistGroup,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    10,
			Block:    2 * time.Second,
		}).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("auth persist: read: %v", err)
			time.Sleep(time.Second)
			continue
		}
		for _, s := range streams {
			for _, msg := range s.Messages {
				w.handle(ctx, stream, msg)
			}
		}
	}
}

func (w *PersistWorker) handle(ctx context.Context, stream string, msg redis.XMessage) {
	raw, _ := msg.Values["job"].(string)
	var job PersistJob
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		w.ack(ctx, stream, msg.ID)
		return
	}
	actor := job.Actor
	if actor == "" {
		actor = auditctx.SystemActor
	}
	actx := auditctx.WithActor(ctx, actor)
	var err error
	switch job.Op {
	case "refresh_create":
		err = w.store.SaveRefreshToken(actx, job.ID, job.UserID, job.TokenHash, job.ExpiresAt)
	case "refresh_revoke":
		err = w.store.RevokeRefreshToken(actx, job.TokenHash)
	default:
		log.Printf("auth persist: unknown op %q", job.Op)
	}
	if err != nil {
		log.Printf("auth persist: %s %s: %v", job.Op, job.TokenHash, err)
		_ = w.cache.rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: w.cache.key(persistDLQ),
			Values: map[string]any{"job": raw, "error": err.Error()},
		}).Err()
	}
	w.ack(ctx, stream, msg.ID)
}

func (w *PersistWorker) ack(ctx context.Context, stream, id string) {
	_ = w.cache.rdb.XAck(ctx, stream, persistGroup, id).Err()
}

func (w *PersistWorker) Lag(ctx context.Context) (int64, error) {
	if w == nil || w.cache == nil {
		return 0, nil
	}
	return w.cache.PersistLag(ctx)
}

func (s *Service) persistRefreshCreate(ctx context.Context, id, userID, hash string, expiresAt time.Time) error {
	if s.writeBehind && s.cache != nil && s.cache.Enabled() {
		row := CachedRefresh{ID: id, UserID: userID, ExpiresAt: expiresAt}
		if err := s.cache.PutRefresh(ctx, hash, row); err != nil {
			return err
		}
		return s.cache.EnqueuePersist(ctx, PersistJob{
			Op:        "refresh_create",
			ID:        id,
			UserID:    userID,
			TokenHash: hash,
			ExpiresAt: expiresAt,
			Actor:     auditctx.ActorID(ctx),
		})
	}
	return s.store.SaveRefreshToken(ctx, id, userID, hash, expiresAt)
}

func (s *Service) persistRefreshRevoke(ctx context.Context, hash string) error {
	if s.writeBehind && s.cache != nil && s.cache.Enabled() {
		_ = s.cache.RevokeRefresh(ctx, hash)
		return s.cache.EnqueuePersist(ctx, PersistJob{
			Op:        "refresh_revoke",
			TokenHash: hash,
			Actor:     auditctx.ActorID(ctx),
		})
	}
	return s.store.RevokeRefreshToken(ctx, hash)
}

func userEmailVerified(user CachedUser) bool {
	if user.AuthProvider != "" && user.AuthProvider != "email" {
		return true
	}
	return user.EmailVerified
}

func cachedFromStore(user store.AuthUser, includeHash bool) CachedUser {
	cu := CachedUser{
		ID:            user.ID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Status:        user.Status,
		Role:          user.Role,
		TenantID:      user.TenantID,
		AuthProvider:  user.AuthProvider,
		EmailVerified: user.EmailVerifiedAt != nil,
	}
	if includeHash {
		cu.PasswordHash = user.PasswordHash
	}
	return cu
}

func profileFromCached(user CachedUser) UserProfile {
	return UserProfile{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        Role(user.Role),
		TenantID:    user.TenantID,
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}