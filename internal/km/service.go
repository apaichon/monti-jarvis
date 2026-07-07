package km

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/scope"
)

type DocumentStore interface {
	CreateKnowledgeDocument(ctx context.Context, doc Document) (Document, error)
	UpdateKnowledgeDocumentStatus(ctx context.Context, id, status string, version, chunkCount int) error
	ListKnowledgeDocuments(ctx context.Context, tenantID, agentID string) ([]Document, error)
	GetKnowledgeDocument(ctx context.Context, id string) (Document, error)
	DeleteAgentKnowledge(ctx context.Context, tenantID, agentID string) ([]string, error)
	ReplaceKnowledgeChunks(ctx context.Context, tenantID, agentID, documentID, kmScope string, chunks []Chunk, chunkIDs []string) error
	CountAgentKnowledge(ctx context.Context, tenantID, agentID string) (docs, chunks int, err error)
	PutKMObject(ctx context.Context, objectKey, contentType string, data []byte) error
	DeleteKMObjects(ctx context.Context, keys []string) error
}

type Service struct {
	store  DocumentStore
	ch     *clickhouse.Client
	embed  *gemini.Client
	tenant string
}

func NewService(store DocumentStore, ch *clickhouse.Client, embed *gemini.Client, tenantID string) *Service {
	return &Service{store: store, ch: ch, embed: embed, tenant: tenantID}
}

func (s *Service) Ingest(ctx context.Context, agentID, filename string, data []byte, kmScope string) (Document, error) {
	agentID = strings.ToLower(strings.TrimSpace(agentID))
	if !scope.ValidAgent(agentID) {
		return Document{}, fmt.Errorf("unknown agent_id %q", agentID)
	}
	if kmScope == "" {
		kmScope = scope.DefaultScope(agentID)
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = "document.md"
	}
	if len(data) == 0 {
		return Document{}, fmt.Errorf("file is empty")
	}

	docID := newID()
	objectKey := path.Join("km", s.tenant, agentID, docID, "original", filename)
	doc := Document{
		ID:        docID,
		TenantID:  s.tenant,
		AgentID:   agentID,
		Filename:  filename,
		ObjectKey: objectKey,
		Mime:      mimeFor(filename),
		Status:    StatusUploaded,
		KMScope:   kmScope,
		KMVersion: 1,
	}
	if err := s.store.PutKMObject(ctx, objectKey, doc.Mime, data); err != nil {
		return Document{}, err
	}
	created, err := s.store.CreateKnowledgeDocument(ctx, doc)
	if err != nil {
		return Document{}, err
	}
	indexed, err := s.indexDocument(ctx, created, string(data))
	if err != nil {
		_ = s.store.UpdateKnowledgeDocumentStatus(ctx, created.ID, StatusFailed, created.KMVersion, 0)
		return created, err
	}
	return indexed, nil
}

func (s *Service) indexDocument(ctx context.Context, doc Document, text string) (Document, error) {
	_ = s.store.UpdateKnowledgeDocumentStatus(ctx, doc.ID, StatusIndexing, doc.KMVersion, 0)
	chunks := ChunkText(text, 0)
	if len(chunks) == 0 {
		return doc, fmt.Errorf("no chunks produced")
	}

	chunkIDs := make([]string, len(chunks))
	rows := make([]clickhouse.EmbeddingRow, 0, len(chunks))
	for i, chunk := range chunks {
		chunkID := newID()
		chunkIDs[i] = chunkID
		vec, err := s.embed.EmbedDocument(ctx, chunk.Content)
		if err != nil {
			return doc, err
		}
		rows = append(rows, clickhouse.EmbeddingRow{
			TenantID:   s.tenant,
			AgentID:    doc.AgentID,
			DocumentID: doc.ID,
			ChunkID:    chunkID,
			KMScope:    doc.KMScope,
			KMVersion:  uint32(doc.KMVersion),
			Content:    chunk.Content,
			Embedding:  vec,
		})
	}

	if err := s.store.ReplaceKnowledgeChunks(ctx, s.tenant, doc.AgentID, doc.ID, doc.KMScope, chunks, chunkIDs); err != nil {
		return doc, err
	}
	if s.ch != nil && s.ch.Enabled() {
		if err := s.ch.ReplaceDocumentEmbeddings(ctx, s.tenant, doc.AgentID, doc.ID, uint32(doc.KMVersion), rows); err != nil {
			return doc, err
		}
	}
	if err := s.store.UpdateKnowledgeDocumentStatus(ctx, doc.ID, StatusIndexed, doc.KMVersion, len(chunks)); err != nil {
		return doc, err
	}
	doc.Status = StatusIndexed
	doc.ChunkCount = len(chunks)
	return doc, nil
}

func (s *Service) ListAgentDocuments(ctx context.Context, agentID string) ([]Document, error) {
	return s.store.ListKnowledgeDocuments(ctx, s.tenant, strings.ToLower(strings.TrimSpace(agentID)))
}

func (s *Service) AgentKnowledge(ctx context.Context, agentID string) (AgentKnowledge, error) {
	agentID = strings.ToLower(strings.TrimSpace(agentID))
	docs, err := s.store.ListKnowledgeDocuments(ctx, s.tenant, agentID)
	if err != nil {
		return AgentKnowledge{}, err
	}
	docCount, chunkCount, err := s.store.CountAgentKnowledge(ctx, s.tenant, agentID)
	if err != nil {
		return AgentKnowledge{}, err
	}
	return AgentKnowledge{
		AgentID:       agentID,
		TenantID:      s.tenant,
		Scope:         scope.DefaultScope(agentID),
		DocumentCount: docCount,
		ChunkCount:    chunkCount,
		Documents:     docs,
	}, nil
}

func (s *Service) ResetAgent(ctx context.Context, agentID string) error {
	agentID = strings.ToLower(strings.TrimSpace(agentID))
	if !scope.ValidAgent(agentID) {
		return fmt.Errorf("unknown agent_id %q", agentID)
	}
	keys, err := s.store.DeleteAgentKnowledge(ctx, s.tenant, agentID)
	if err != nil {
		return err
	}
	_ = s.store.DeleteKMObjects(ctx, keys)
	if s.ch != nil && s.ch.Enabled() {
		return s.ch.DeleteAgentEmbeddings(ctx, s.tenant, agentID)
	}
	return nil
}

func mimeFor(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".md"):
		return "text/markdown"
	case strings.HasSuffix(lower, ".txt"):
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}