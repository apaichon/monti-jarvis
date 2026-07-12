package clickhouse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL  string
	db       string
	user     string
	password string
	http     *http.Client
}

type ChunkHit struct {
	ChunkID    string
	DocumentID string
	AgentID    string
	KMScope    string
	Content    string
	Score      float64
}

func New(baseURL, database, user, password string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	if database == "" {
		database = "monti_jarvis"
	}
	return &Client{
		baseURL:  baseURL,
		db:       database,
		user:     strings.TrimSpace(user),
		password: password,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.baseURL != ""
}

func (c *Client) Ping(ctx context.Context) error {
	if !c.Enabled() {
		return fmt.Errorf("clickhouse is not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/ping", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("clickhouse ping: http %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) EnsureSchema(ctx context.Context) error {
	db := quoteIdent(c.db)
	statements := []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", db),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.km_embeddings (
  tenant_id String,
  agent_id String,
  document_id String,
  chunk_id String,
  km_scope String,
  km_version UInt32,
  content String,
  embedding Array(Float32),
  created_at DateTime DEFAULT now(),
  updated_at DateTime DEFAULT now(),
  created_by String DEFAULT 'system',
  updated_by String DEFAULT 'system'
) ENGINE = MergeTree()
ORDER BY (tenant_id, agent_id, km_scope, chunk_id)`, db),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.qa_events (
  event_id String,
  tenant_id String,
  agent_id String,
  topic String,
  question String,
  event_type String,
  created_at DateTime DEFAULT now(),
  updated_at DateTime DEFAULT now(),
  created_by String DEFAULT 'system',
  updated_by String DEFAULT 'system'
) ENGINE = MergeTree()
ORDER BY (tenant_id, created_at)`, db),
	}
	for _, stmt := range statements {
		if err := c.exec(ctx, stmt); err != nil {
			return err
		}
	}
	return c.ensureAuditSchema(ctx)
}

func (c *Client) ensureAuditSchema(ctx context.Context) error {
	if !c.Enabled() {
		return nil
	}
	db := quoteIdent(c.db)
	stmts := []string{
		fmt.Sprintf(`ALTER TABLE %s.km_embeddings
  ADD COLUMN IF NOT EXISTS created_at DateTime DEFAULT now(),
  ADD COLUMN IF NOT EXISTS created_by String DEFAULT 'system',
  ADD COLUMN IF NOT EXISTS updated_by String DEFAULT 'system'`, db),
		fmt.Sprintf(`ALTER TABLE %s.qa_events
  ADD COLUMN IF NOT EXISTS updated_at DateTime DEFAULT now(),
  ADD COLUMN IF NOT EXISTS created_by String DEFAULT 'system',
  ADD COLUMN IF NOT EXISTS updated_by String DEFAULT 'system'`, db),
	}
	for _, stmt := range stmts {
		if err := c.exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

type EmbeddingRow struct {
	TenantID   string
	AgentID    string
	DocumentID string
	ChunkID    string
	KMScope    string
	KMVersion  uint32
	Content    string
	Embedding  []float32
}

func (c *Client) ReplaceDocumentEmbeddings(ctx context.Context, tenantID, agentID, documentID string, version uint32, rows []EmbeddingRow) error {
	if !c.Enabled() {
		return fmt.Errorf("clickhouse is not configured")
	}
	del := fmt.Sprintf(
		"ALTER TABLE %s.km_embeddings DELETE WHERE tenant_id = '%s' AND agent_id = '%s' AND document_id = '%s'",
		quoteIdent(c.db), escape(tenantID), escape(agentID), escape(documentID),
	)
	if err := c.exec(ctx, del); err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	return c.insertEmbeddings(ctx, rows)
}

func (c *Client) DeleteAgentEmbeddings(ctx context.Context, tenantID, agentID string) error {
	if !c.Enabled() {
		return fmt.Errorf("clickhouse is not configured")
	}
	q := fmt.Sprintf(
		"ALTER TABLE %s.km_embeddings DELETE WHERE tenant_id = '%s' AND agent_id = '%s'",
		quoteIdent(c.db), escape(tenantID), escape(agentID),
	)
	return c.exec(ctx, q)
}

// DeleteDocumentEmbeddings removes vectors for one document.
func (c *Client) DeleteDocumentEmbeddings(ctx context.Context, tenantID, agentID, documentID string) error {
	if !c.Enabled() {
		return nil
	}
	q := fmt.Sprintf(
		"ALTER TABLE %s.km_embeddings DELETE WHERE tenant_id = '%s' AND agent_id = '%s' AND document_id = '%s'",
		quoteIdent(c.db), escape(tenantID), escape(agentID), escape(documentID),
	)
	return c.exec(ctx, q)
}

// UpdateDocumentScope retags embeddings' km_scope for a document (no re-embed).
func (c *Client) UpdateDocumentScope(ctx context.Context, tenantID, agentID, documentID, kmScope string) error {
	if !c.Enabled() {
		return nil
	}
	q := fmt.Sprintf(
		"ALTER TABLE %s.km_embeddings UPDATE km_scope = '%s' WHERE tenant_id = '%s' AND agent_id = '%s' AND document_id = '%s'",
		quoteIdent(c.db), escape(kmScope), escape(tenantID), escape(agentID), escape(documentID),
	)
	return c.exec(ctx, q)
}

func (c *Client) InsertQAEvent(ctx context.Context, tenantID, agentID, topic, question, eventType string) error {
	if !c.Enabled() {
		return nil
	}
	q := fmt.Sprintf(
		`INSERT INTO %s.qa_events (event_id, tenant_id, agent_id, topic, question, event_type) VALUES ('%s','%s','%s','%s','%s','%s')`,
		quoteIdent(c.db),
		escape(randomID()),
		escape(tenantID),
		escape(agentID),
		escape(topic),
		escape(question),
		escape(eventType),
	)
	return c.exec(ctx, q)
}

func (c *Client) Search(ctx context.Context, tenantID, agentID string, scopes []string, query []float32, topK, candidateLimit int) ([]ChunkHit, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("clickhouse is not configured")
	}
	if topK <= 0 {
		topK = 5
	}
	if candidateLimit <= 0 {
		candidateLimit = 50
	}
	scopeFilter := ""
	if len(scopes) > 0 {
		parts := make([]string, len(scopes))
		for i, s := range scopes {
			parts[i] = "'" + escape(s) + "'"
		}
		scopeFilter = " AND km_scope IN (" + strings.Join(parts, ",") + ")"
	}

	agentFilter := ""
	if strings.TrimSpace(agentID) != "" {
		agentFilter = " AND agent_id = '" + escape(agentID) + "'"
	}

	// Fetch candidate rows and rank in Go for portable dev setups.
	q := fmt.Sprintf(`
SELECT chunk_id, document_id, agent_id, km_scope, content, embedding
FROM %s.km_embeddings
WHERE tenant_id = '%s'%s%s
LIMIT %d
FORMAT JSON`,
		quoteIdent(c.db), escape(tenantID), agentFilter, scopeFilter, candidateLimit,
	)

	body, err := c.query(ctx, q)
	if err != nil {
		return nil, err
	}
	type row struct {
		ChunkID    string    `json:"chunk_id"`
		DocumentID string    `json:"document_id"`
		AgentID    string    `json:"agent_id"`
		KMScope    string    `json:"km_scope"`
		Content    string    `json:"content"`
		Embedding  []float32 `json:"embedding"`
	}
	var parsed struct {
		Data []row `json:"data"`
	}
	if err := jsonUnmarshal(body, &parsed); err != nil {
		return nil, err
	}

	hits := make([]ChunkHit, 0, len(parsed.Data))
	for _, r := range parsed.Data {
		hits = append(hits, ChunkHit{
			ChunkID:    r.ChunkID,
			DocumentID: r.DocumentID,
			AgentID:    r.AgentID,
			KMScope:    r.KMScope,
			Content:    r.Content,
			Score:      cosineSimilarity(query, r.Embedding),
		})
	}
	sortHits(hits)
	if len(hits) > topK {
		hits = hits[:topK]
	}
	return hits, nil
}

func (c *Client) ListAgentChunks(ctx context.Context, tenantID, agentID string, scopes []string, limit int) ([]ChunkHit, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("clickhouse is not configured")
	}
	if limit <= 0 {
		limit = 8
	}
	scopeFilter := ""
	if len(scopes) > 0 {
		parts := make([]string, len(scopes))
		for i, s := range scopes {
			parts[i] = "'" + escape(s) + "'"
		}
		scopeFilter = " AND km_scope IN (" + strings.Join(parts, ",") + ")"
	}
	agentFilter := ""
	if strings.TrimSpace(agentID) != "" {
		agentFilter = " AND agent_id = '" + escape(agentID) + "'"
	}
	q := fmt.Sprintf(`
SELECT chunk_id, document_id, agent_id, km_scope, content
FROM %s.km_embeddings
WHERE tenant_id = '%s'%s%s
ORDER BY updated_at DESC
LIMIT %d
FORMAT JSON`,
		quoteIdent(c.db), escape(tenantID), agentFilter, scopeFilter, limit,
	)
	body, err := c.query(ctx, q)
	if err != nil {
		return nil, err
	}
	type row struct {
		ChunkID    string `json:"chunk_id"`
		DocumentID string `json:"document_id"`
		AgentID    string `json:"agent_id"`
		KMScope    string `json:"km_scope"`
		Content    string `json:"content"`
	}
	var parsed struct {
		Data []row `json:"data"`
	}
	if err := jsonUnmarshal(body, &parsed); err != nil {
		return nil, err
	}
	out := make([]ChunkHit, 0, len(parsed.Data))
	for _, r := range parsed.Data {
		out = append(out, ChunkHit{
			ChunkID:    r.ChunkID,
			DocumentID: r.DocumentID,
			AgentID:    r.AgentID,
			KMScope:    r.KMScope,
			Content:    r.Content,
			Score:      1,
		})
	}
	return out, nil
}

// insertEmbeddings bulk-loads rows using JSONEachRow so markdown content with
// quotes/newlines cannot break SQL VALUES parsing (CH 400 CANNOT_PARSE_INPUT).
func (c *Client) insertEmbeddings(ctx context.Context, rows []EmbeddingRow) error {
	if len(rows) == 0 {
		return nil
	}
	type payload struct {
		TenantID   string    `json:"tenant_id"`
		AgentID    string    `json:"agent_id"`
		DocumentID string    `json:"document_id"`
		ChunkID    string    `json:"chunk_id"`
		KMScope    string    `json:"km_scope"`
		KMVersion  uint32    `json:"km_version"`
		Content    string    `json:"content"`
		Embedding  []float32 `json:"embedding"`
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// ClickHouse JSONEachRow is one object per line; Encoder adds newlines.
	for _, row := range rows {
		if err := enc.Encode(payload{
			TenantID:   row.TenantID,
			AgentID:    row.AgentID,
			DocumentID: row.DocumentID,
			ChunkID:    row.ChunkID,
			KMScope:    row.KMScope,
			KMVersion:  row.KMVersion,
			Content:    row.Content,
			Embedding:  row.Embedding,
		}); err != nil {
			return fmt.Errorf("clickhouse encode row: %w", err)
		}
	}
	// Query in URL param; body is pure JSONEachRow data (safe for large text).
	q := fmt.Sprintf(
		"INSERT INTO %s.km_embeddings (tenant_id, agent_id, document_id, chunk_id, km_scope, km_version, content, embedding) FORMAT JSONEachRow",
		quoteIdent(c.db),
	)
	return c.execInsert(ctx, q, buf.Bytes())
}

func (c *Client) exec(ctx context.Context, query string) error {
	_, err := c.query(ctx, query)
	return err
}

// execInsert runs an INSERT with body as the data payload (FORMAT in query string).
func (c *Client) execInsert(ctx context.Context, query string, data []byte) error {
	params := url.Values{"database": {c.db}, "query": {query}}
	if c.user != "" {
		params.Set("user", c.user)
	}
	if c.password != "" {
		params.Set("password", c.password)
	}
	endpoint := c.baseURL + "/?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("clickhouse: http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) query(ctx context.Context, query string) ([]byte, error) {
	params := url.Values{"database": {c.db}}
	if c.user != "" {
		params.Set("user", c.user)
	}
	if c.password != "" {
		params.Set("password", c.password)
	}
	endpoint := c.baseURL + "/?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(query))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("clickhouse: http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func quoteIdent(name string) string {
	return name
}

func escape(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}