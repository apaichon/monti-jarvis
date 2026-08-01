package quota

import (
	"time"

	"github.com/libra/monti-jarvis/internal/store"
)

// PackageSummary mirrors entitlements package fields for snapshots.
type PackageSummary struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Limits is the numeric/bool ceiling from the active package rules snapshot.
type Limits struct {
	MaxAIEmployees        int  `json:"max_ai_employees"`
	MaxMonthlyCallMinutes int  `json:"max_monthly_call_minutes"`
	MaxMobileCallMinutes  int  `json:"max_mobile_call_minutes"`
	MaxKMDocuments        int  `json:"max_km_documents"`
	MaxStorageBytes       int  `json:"max_storage_bytes"`
	MaxConcurrentCalls    int  `json:"max_concurrent_calls"`
	VoiceEnabled          bool `json:"voice_enabled"`
	RAGEnabled            bool `json:"rag_enabled"`
}

// Usage is current consumption for Snapshot.
type Usage struct {
	AIEmployees        int `json:"ai_employees"`
	MonthlyCallMinutes int `json:"monthly_call_minutes"`
	MobileCallMinutes  int `json:"mobile_call_minutes"`
	KMDocuments        int `json:"km_documents"`
	StorageBytes       int `json:"storage_bytes"`
	ConcurrentCalls    int `json:"concurrent_calls"`
}

// ConcurrentQueueSnapshot reports live package-call capacity for support and UI.
type ConcurrentQueueSnapshot struct {
	QueueEnabled       bool   `json:"queue_enabled"`
	ActiveCalls        int    `json:"active_calls"`
	QueuedCallers      int    `json:"queued_callers"`
	TotalCalls         int    `json:"total_calls"`
	MaxConcurrentCalls int    `json:"max_concurrent_calls"`
	BusyStatus         string `json:"busy_status"`
	OldestWaitSeconds  int    `json:"oldest_wait_seconds"`
	RecentTimeouts24h  int    `json:"recent_timeouts_24h"`
}

// QueueUpdate is sent while a caller waits for a concurrent-call slot.
type QueueUpdate struct {
	Type                 string
	AdmissionID          string
	Position             int
	EstimatedWaitSeconds int
	Snapshot             ConcurrentQueueSnapshot
	Message              string
}

// QueuedAdmission is returned after a caller owns a concurrent-call slot.
type QueuedAdmission struct {
	AdmissionID string
	Release     func()
	Snapshot    ConcurrentQueueSnapshot
}

// Dimension is the stable reporting shape shared by tenant, platform, and
// mobile clients. Source/freshness are explicit so unavailable dependencies
// are never rendered as an authoritative zero.
type Dimension struct {
	Dimension      string     `json:"dimension"`
	Unit           string     `json:"unit"`
	Period         string     `json:"period"`
	Limit          int        `json:"limit"` // total limit, retained for existing clients
	BaseLimit      int        `json:"base_limit"`
	BonusGranted   int        `json:"bonus_granted"`
	BonusUsed      int        `json:"bonus_used"`
	BonusRemaining int        `json:"bonus_remaining"`
	TotalLimit     int        `json:"total_limit"`
	BonusExpiresAt *time.Time `json:"bonus_expires_at,omitempty"`
	Consumed       *int       `json:"consumed"`
	Remaining      *int       `json:"remaining"`
	Source         string     `json:"source"`
	Freshness      string     `json:"freshness"`
}

// Snapshot is returned by Service.Snapshot for platform admin UI.
type Snapshot struct {
	TenantID        string                   `json:"tenant_id"`
	Package         *PackageSummary          `json:"package"`
	Status          string                   `json:"status"` // active | none
	Period          string                   `json:"period"` // YYYY-MM UTC
	Limits          *Limits                  `json:"limits"`
	Usage           Usage                    `json:"usage"`
	Bonus           []store.BonusBalance     `json:"bonus"`
	Dimensions      []Dimension              `json:"current_dimensions"`
	ConcurrentQueue *ConcurrentQueueSnapshot `json:"concurrent_queue,omitempty"`
}

// Rule dimension keys (package rules-v1).
const (
	DimMaxAIEmployees        = "max_ai_employees"
	DimMaxMonthlyCallMinutes = "max_monthly_call_minutes"
	DimMaxMobileCallMinutes  = "max_mobile_call_minutes"
	DimMaxKMDocuments        = "max_km_documents"
	DimMaxStorageBytes       = "max_storage_bytes"
	DimMaxConcurrentCalls    = "max_concurrent_calls"
	DimVoiceEnabled          = "voice_enabled"
	DimRAGEnabled            = "rag_enabled"
)

// Rate buckets for AllowRate.
const (
	BucketChat  = "chat"
	BucketKM    = "km"
	BucketVoice = "voice"
)

const (
	UnitMinutes     = "minutes"
	UnitDocuments   = "documents"
	UnitBytes       = "bytes"
	UnitAssignments = "assignments"
	UnitCalls       = "calls"
)
