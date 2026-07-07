package km

import "time"

type Document struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	AgentID   string    `json:"agent_id"`
	Filename  string    `json:"filename"`
	ObjectKey string    `json:"object_key"`
	Mime      string    `json:"mime"`
	Status    string    `json:"status"`
	KMScope   string    `json:"km_scope"`
	KMVersion int       `json:"km_version"`
	ChunkCount int      `json:"chunk_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AgentKnowledge struct {
	AgentID      string     `json:"agent_id"`
	TenantID     string     `json:"tenant_id"`
	Scope        string     `json:"scope"`
	DocumentCount int       `json:"document_count"`
	ChunkCount   int        `json:"chunk_count"`
	Documents    []Document `json:"documents"`
}

const (
	StatusUploaded = "uploaded"
	StatusIndexing = "indexing"
	StatusIndexed  = "indexed"
	StatusFailed   = "failed"
)