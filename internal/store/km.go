package store

import (
	"bytes"
	"context"
	"fmt"

	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/libra/monti-jarvis/internal/km"
	"github.com/minio/minio-go/v7"
)

func (s *Store) CreateKnowledgeDocument(ctx context.Context, doc km.Document) (km.Document, error) {
	if s.pg == nil {
		return km.Document{}, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO %s.knowledge_documents
  (id, tenant_id, agent_id, filename, object_key, mime, status, km_scope, km_version, created_by, updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)
RETURNING id, tenant_id, agent_id, filename, object_key, mime, status, km_scope, km_version, created_at, updated_at`,
		schema),
		doc.ID, doc.TenantID, doc.AgentID, doc.Filename, doc.ObjectKey, doc.Mime, doc.Status, doc.KMScope, doc.KMVersion, actor,
	).Scan(
		&doc.ID, &doc.TenantID, &doc.AgentID, &doc.Filename, &doc.ObjectKey, &doc.Mime, &doc.Status, &doc.KMScope, &doc.KMVersion, &doc.CreatedAt, &doc.UpdatedAt,
	)
	return doc, err
}

func (s *Store) UpdateKnowledgeDocumentStatus(ctx context.Context, id, status string, version, chunkCount int) error {
	if s.pg == nil {
		return fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.knowledge_documents
SET status = $2, km_version = $3, chunk_count = $4, updated_by = $5
WHERE id = $1`, schema), id, status, version, chunkCount, actor)
	return err
}

func (s *Store) ListKnowledgeDocuments(ctx context.Context, tenantID, agentID string) ([]km.Document, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT id, tenant_id, agent_id, filename, object_key, mime, status, km_scope, km_version, chunk_count, created_at, updated_at
FROM %s.knowledge_documents
WHERE tenant_id = $1 AND agent_id = $2
ORDER BY created_at DESC`, schema), tenantID, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []km.Document
	for rows.Next() {
		var doc km.Document
		if err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.AgentID, &doc.Filename, &doc.ObjectKey, &doc.Mime, &doc.Status, &doc.KMScope, &doc.KMVersion, &doc.ChunkCount, &doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, rows.Err()
}

func (s *Store) GetKnowledgeDocument(ctx context.Context, id string) (km.Document, error) {
	if s.pg == nil {
		return km.Document{}, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var doc km.Document
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, tenant_id, agent_id, filename, object_key, mime, status, km_scope, km_version, chunk_count, created_at, updated_at
FROM %s.knowledge_documents WHERE id = $1`, schema), id,
	).Scan(
		&doc.ID, &doc.TenantID, &doc.AgentID, &doc.Filename, &doc.ObjectKey, &doc.Mime, &doc.Status, &doc.KMScope, &doc.KMVersion, &doc.ChunkCount, &doc.CreatedAt, &doc.UpdatedAt,
	)
	return doc, err
}

func (s *Store) DeleteAgentKnowledge(ctx context.Context, tenantID, agentID string) ([]string, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT object_key FROM %s.knowledge_documents WHERE tenant_id = $1 AND agent_id = $2`, schema), tenantID, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if _, err := s.pg.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s.knowledge_chunks WHERE tenant_id = $1 AND agent_id = $2`, schema), tenantID, agentID); err != nil {
		return nil, err
	}
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s.knowledge_documents WHERE tenant_id = $1 AND agent_id = $2`, schema), tenantID, agentID)
	return keys, err
}

func (s *Store) ReplaceKnowledgeChunks(ctx context.Context, tenantID, agentID, documentID, kmScope string, chunks []km.Chunk, chunkIDs []string) error {
	if s.pg == nil {
		return fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s.knowledge_chunks WHERE document_id = $1`, schema), documentID); err != nil {
		return err
	}
	actor := auditctx.ActorID(ctx)
	for i, chunk := range chunks {
		chunkID := chunkIDs[i]
		if _, err := tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.knowledge_chunks
  (id, document_id, tenant_id, agent_id, chunk_index, content, km_scope, created_by, updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)`, schema),
			chunkID, documentID, tenantID, agentID, chunk.Index, chunk.Content, kmScope, actor,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) CountAgentKnowledge(ctx context.Context, tenantID, agentID string) (docs, chunks int, err error) {
	if s.pg == nil {
		return 0, 0, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	err = s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s.knowledge_documents WHERE tenant_id = $1 AND agent_id = $2`, schema), tenantID, agentID).Scan(&docs)
	if err != nil {
		return 0, 0, err
	}
	err = s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s.knowledge_chunks WHERE tenant_id = $1 AND agent_id = $2`, schema), tenantID, agentID).Scan(&chunks)
	return docs, chunks, err
}

// CountTenantKnowledgeDocuments returns total KM documents for a tenant (all agents).
// Used by SPRINT-013 quota for max_km_documents.
func (s *Store) CountTenantKnowledgeDocuments(ctx context.Context, tenantID string) (int, error) {
	if s.pg == nil {
		return 0, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var n int
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s.knowledge_documents WHERE tenant_id = $1`, schema), tenantID).Scan(&n)
	return n, err
}

func (s *Store) PutKMObject(ctx context.Context, objectKey, contentType string, data []byte) error {
	if s.minio == nil {
		return fmt.Errorf("minio is not available")
	}
	_, err := s.minio.PutObject(ctx, s.cfg.MinioBucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *Store) DeleteKMObjects(ctx context.Context, keys []string) error {
	if s.minio == nil {
		return fmt.Errorf("minio is not available")
	}
	for _, key := range keys {
		_ = s.minio.RemoveObject(ctx, s.cfg.MinioBucket, key, minio.RemoveObjectOptions{})
	}
	return nil
}