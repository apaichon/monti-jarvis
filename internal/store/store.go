package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/libra/monti-jarvis/internal/env"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	cfg   env.Config
	pg    *pgxpool.Pool
	redis *redis.Client
	minio *minio.Client
}

type Health struct {
	Postgres   string `json:"postgres"`
	Redis      string `json:"redis"`
	Minio      string `json:"minio"`
	ClickHouse string `json:"clickhouse"`
	NATS       string `json:"nats"`
	LiveKit    string `json:"livekit"`
}

func Open(ctx context.Context, cfg env.Config) (*Store, []string) {
	s := &Store{cfg: cfg}
	var warnings []string

	if cfg.PostgresURL != "" {
		pool, err := pgxpool.New(ctx, cfg.PostgresURL)
		if err != nil {
			warnings = append(warnings, "postgres config: "+err.Error())
		} else if err := pool.Ping(ctx); err != nil {
			warnings = append(warnings, "postgres ping: "+err.Error())
			pool.Close()
		} else {
			s.pg = pool
			if err := s.ensureSchema(ctx); err != nil {
				warnings = append(warnings, "postgres schema: "+err.Error())
			} else if err := s.SeedPaymentGatewayFromEnv(ctx); err != nil {
				warnings = append(warnings, "payment gateway seed: "+err.Error())
			}
		}
	}

	if cfg.RedisURL != "" {
		opts, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			warnings = append(warnings, "redis config: "+err.Error())
		} else {
			client := redis.NewClient(opts)
			if err := client.Ping(ctx).Err(); err != nil {
				warnings = append(warnings, "redis ping: "+err.Error())
				_ = client.Close()
			} else {
				s.redis = client
			}
		}
	}

	if cfg.MinioEndpoint != "" {
		client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
			Secure: cfg.MinioUseSSL,
		})
		if err != nil {
			warnings = append(warnings, "minio config: "+err.Error())
		} else {
			s.minio = client
		}
	}

	return s, warnings
}

func (s *Store) Redis() *redis.Client {
	if s == nil {
		return nil
	}
	return s.redis
}

func (s *Store) Close() {
	if s.pg != nil {
		s.pg.Close()
	}
	if s.redis != nil {
		_ = s.redis.Close()
	}
}

func (s *Store) Health(ctx context.Context) Health {
	h := Health{Postgres: "disabled", Redis: "disabled", Minio: "disabled", ClickHouse: "disabled"}
	if s.pg != nil {
		h.Postgres = "ok"
		if err := s.pg.Ping(ctx); err != nil {
			h.Postgres = err.Error()
		}
	}
	if s.redis != nil {
		h.Redis = "ok"
		if err := s.redis.Ping(ctx).Err(); err != nil {
			h.Redis = err.Error()
		}
	}
	if s.minio != nil {
		h.Minio = "ok"
		exists, err := s.minio.BucketExists(ctx, s.cfg.MinioBucket)
		if err != nil {
			h.Minio = err.Error()
		} else if !exists {
			h.Minio = "bucket missing: " + s.cfg.MinioBucket
		}
	}
	return h
}

func (s *Store) SaveExchange(ctx context.Context, sessionID, agentID, userText, assistantText string) {
	if s.pg != nil {
		actor := auditctx.ActorID(ctx)
		schema := quoteIdent(s.cfg.PostgresSchema)
		_, _ = s.pg.Exec(ctx,
			fmt.Sprintf(`INSERT INTO %s.calls (id, agent_id, created_by, updated_by)
VALUES ($1, $2, $3, $3)
ON CONFLICT (id) DO UPDATE SET agent_id = EXCLUDED.agent_id, updated_by = EXCLUDED.updated_by`, schema),
			sessionID, agentID, actor,
		)
		_, _ = s.pg.Exec(ctx,
			fmt.Sprintf(`INSERT INTO %s.messages (call_id, role, content, created_by, updated_by)
VALUES ($1, 'caller', $2, $3, $3), ($1, 'agent', $4, $3, $3)`, schema),
			sessionID, userText, actor, assistantText,
		)
	}
	if s.redis != nil {
		key := s.cfg.RedisPrefix + "call:" + sessionID
		pipe := s.redis.Pipeline()
		pipe.HSet(ctx, key, "updated_at", time.Now().UTC().Format(time.RFC3339), "agent_id", agentID)
		pipe.Expire(ctx, key, 24*time.Hour)
		_, _ = pipe.Exec(ctx)
	}
}

func (s *Store) ensureSchema(ctx context.Context) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE SCHEMA IF NOT EXISTS %s;
CREATE TABLE IF NOT EXISTS %s.calls (
  id text PRIMARY KEY,
  agent_id text NOT NULL DEFAULT 'ava',
  title text NOT NULL DEFAULT 'Inbound call',%s
);
CREATE TABLE IF NOT EXISTS %s.messages (
  id bigserial PRIMARY KEY,
  call_id text NOT NULL REFERENCES %s.calls(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('caller', 'agent')),
  content text NOT NULL,%s
);
CREATE TABLE IF NOT EXISTS %s.call_sessions (
  id text PRIMARY KEY,
  tenant_id text NOT NULL DEFAULT 'demo',
  room_name text NOT NULL UNIQUE,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'ended')),
  started_at timestamptz NOT NULL DEFAULT now(),
  ended_at timestamptz,
  recording_key text,%s
);
CREATE TABLE IF NOT EXISTS %s.call_turns (
  id bigserial PRIMARY KEY,
  call_id text NOT NULL REFERENCES %s.call_sessions(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('caller', 'agent', 'system')),
  content text NOT NULL,
  source_chunk_ids jsonb,%s
);
CREATE TABLE IF NOT EXISTS %s.knowledge_documents (
  id text PRIMARY KEY,
  tenant_id text NOT NULL,
  agent_id text NOT NULL,
  filename text NOT NULL,
  object_key text NOT NULL,
  mime text NOT NULL DEFAULT 'text/plain',
  status text NOT NULL DEFAULT 'uploaded',
  km_scope text NOT NULL DEFAULT 'general',
  km_version integer NOT NULL DEFAULT 1,
  chunk_count integer NOT NULL DEFAULT 0,%s
);
CREATE TABLE IF NOT EXISTS %s.knowledge_chunks (
  id text PRIMARY KEY,
  document_id text NOT NULL REFERENCES %s.knowledge_documents(id) ON DELETE CASCADE,
  tenant_id text NOT NULL,
  agent_id text NOT NULL,
  chunk_index integer NOT NULL,
  content text NOT NULL,
  km_scope text NOT NULL,%s
);
CREATE INDEX IF NOT EXISTS knowledge_documents_agent_idx ON %s.knowledge_documents (tenant_id, agent_id);
CREATE INDEX IF NOT EXISTS knowledge_chunks_agent_idx ON %s.knowledge_chunks (tenant_id, agent_id);`,
		schema, schema, auditColumnsDDL, schema, schema, auditColumnsDDL, schema, auditColumnsDDL, schema, schema, auditColumnsDDL, schema, auditColumnsDDL, schema, schema, auditColumnsDDL, schema, schema))
	if err != nil {
		return err
	}
	if err := s.ensureAuthSchema(ctx); err != nil {
		return err
	}
	if err := s.ensurePackagesSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureAvatarsSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureTenantRegisterSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureTenantAuthSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureTenantKYCSchema(ctx); err != nil {
		return err
	}
	if err := s.ensurePaymentSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureEmbedSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureKMGapsSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureSettingsSchema(ctx); err != nil {
		return err
	}
	if err := s.ensurePreviewSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureTiersSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureCustomersSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureCustomerAuthSchema(ctx); err != nil {
		return err
	}
	if err := s.ensureConversationRecordsSchema(ctx); err != nil {
		return err
	}
	return s.ensureAuditSchema(ctx)
}

func quoteIdent(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "callcenter"
	}
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
