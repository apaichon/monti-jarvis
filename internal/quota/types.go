package quota

// PackageSummary mirrors entitlements package fields for snapshots.
type PackageSummary struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Limits is the numeric/bool ceiling from package rules-v1.
type Limits struct {
	MaxAIEmployees         int  `json:"max_ai_employees"`
	MaxMonthlyCallMinutes  int  `json:"max_monthly_call_minutes"`
	MaxKMDocuments         int  `json:"max_km_documents"`
	MaxConcurrentCalls     int  `json:"max_concurrent_calls"`
	VoiceEnabled           bool `json:"voice_enabled"`
	RAGEnabled             bool `json:"rag_enabled"`
}

// Usage is current consumption for Snapshot.
type Usage struct {
	AIEmployees         int `json:"ai_employees"`
	MonthlyCallMinutes  int `json:"monthly_call_minutes"`
	KMDocuments         int `json:"km_documents"`
	ConcurrentCalls     int `json:"concurrent_calls"`
}

// Snapshot is returned by Service.Snapshot for platform admin UI.
type Snapshot struct {
	TenantID string          `json:"tenant_id"`
	Package  *PackageSummary `json:"package"`
	Status   string          `json:"status"` // active | none
	Period   string          `json:"period"` // YYYY-MM UTC
	Limits   *Limits         `json:"limits"`
	Usage    Usage           `json:"usage"`
}

// Rule dimension keys (package rules-v1).
const (
	DimMaxAIEmployees        = "max_ai_employees"
	DimMaxMonthlyCallMinutes = "max_monthly_call_minutes"
	DimMaxKMDocuments        = "max_km_documents"
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
