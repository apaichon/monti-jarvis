package leads

import (
	"fmt"
	"sort"
	"strings"
)

// PublicPackage is the marketing-safe package projection (no cost internals).
type PublicPackage struct {
	ID             string         `json:"id"`
	Slug           string         `json:"slug,omitempty"`
	Name           string         `json:"name"`
	Description    string         `json:"description,omitempty"`
	PriceAmount    float64        `json:"price_amount"`
	PriceCurrency  string         `json:"price_currency"`
	BillingPeriod  string         `json:"billing_period"`
	PurchaseMode   string         `json:"purchase_mode"` // self_serve | quote
	Deployment     string         `json:"deployment"`    // shared_cloud | dedicated_vm
	Highlights     []string       `json:"highlights"`
	RulesSummary   map[string]any `json:"rules_summary"`
}

// PackageSource is the minimal package fields needed for public projection.
type PackageSource struct {
	ID            string
	Slug          string
	Name          string
	Description   string
	PriceCents    int
	Currency      string
	BillingPeriod string
	Status        string
	Rules         map[string]any
	PurchaseMode  string
	Deployment    string
}

// ProjectPublicPackages maps active packages to public DTOs.
func ProjectPublicPackages(pkgs []PackageSource) []PublicPackage {
	out := make([]PublicPackage, 0, len(pkgs))
	for _, p := range pkgs {
		if strings.ToLower(strings.TrimSpace(p.Status)) != "active" {
			continue
		}
		out = append(out, ProjectPublicPackage(p))
	}
	return out
}

// ProjectPublicPackage maps one package to the public marketing DTO.
func ProjectPublicPackage(p PackageSource) PublicPackage {
	currency := strings.TrimSpace(p.Currency)
	if currency == "" {
		currency = "USD"
	}
	period := normalizeBillingPeriod(p.BillingPeriod)
	summary := rulesSummary(p.Rules)
	mode := strings.TrimSpace(p.PurchaseMode)
	if mode == "" {
		mode = "self_serve"
	}
	deploy := strings.TrimSpace(p.Deployment)
	if deploy == "" {
		deploy = "shared_cloud"
	}
	return PublicPackage{
		ID:            p.ID,
		Slug:          p.Slug,
		Name:          p.Name,
		Description:   p.Description,
		PriceAmount:   float64(p.PriceCents) / 100.0,
		PriceCurrency: currency,
		BillingPeriod: period,
		PurchaseMode:  mode,
		Deployment:    deploy,
		Highlights:    highlightsFromRules(p.Rules),
		RulesSummary:  summary,
	}
}

func normalizeBillingPeriod(period string) string {
	period = strings.ToLower(strings.TrimSpace(period))
	switch period {
	case "monthly", "month", "mo":
		return "month"
	case "yearly", "year", "annual", "annually":
		return "year"
	case "":
		return "month"
	default:
		return period
	}
}

// Map internal rule keys to public-safe summary keys.
var ruleSummaryKeys = map[string]string{
	"max_ai_employees":         "ai_employees",
	"ai_employees":             "ai_employees",
	"max_km_documents":         "km_documents",
	"km_documents":             "km_documents",
	"max_monthly_call_minutes": "monthly_call_minutes",
	"monthly_call_minutes":     "monthly_call_minutes",
	"max_concurrent_calls":     "concurrent_calls",
	"concurrent_calls":         "concurrent_calls",
	"voice_enabled":            "voice_enabled",
	"rag_enabled":              "rag_enabled",
	"max_storage_bytes":        "storage_bytes",
	"max_storage_gb":           "storage_gb",
	"storage_gb":               "storage_gb",
}

func rulesSummary(rules map[string]any) map[string]any {
	out := map[string]any{}
	if rules == nil {
		return out
	}
	for k, v := range rules {
		if public, ok := ruleSummaryKeys[k]; ok {
			// Public marketing: KM count is unlimited when sentinel ≥ 1e6.
			if public == "km_documents" {
				if n, ok := asInt(v); ok && n >= 1_000_000 {
					out["km_documents"] = "unlimited"
					continue
				}
			}
			out[public] = v
		}
	}
	return out
}

func highlightsFromRules(rules map[string]any) []string {
	if rules == nil {
		return []string{}
	}
	type item struct {
		order int
		text  string
	}
	var items []item
	if n, ok := asInt(rules["max_ai_employees"]); ok {
		items = append(items, item{1, fmt.Sprintf("%d AI avatars", n)})
	} else if n, ok := asInt(rules["ai_employees"]); ok {
		items = append(items, item{1, fmt.Sprintf("%d AI avatars", n)})
	}
	// KM document count is commercially unlimited; storage is the real cap.
	storageInKMLine := false
	if n, ok := asInt(rules["max_km_documents"]); ok {
		if n >= 1_000_000 {
			if sb, ok := asInt(rules["max_storage_bytes"]); ok && sb > 0 {
				gb := sb / (1 << 30)
				if gb < 1 {
					gb = 1
				}
				items = append(items, item{2, fmt.Sprintf("Unlimited KM (up to %d GB storage)", gb)})
				storageInKMLine = true
			} else {
				items = append(items, item{2, "Unlimited KM (within storage)"})
			}
		} else {
			items = append(items, item{2, fmt.Sprintf("%d knowledge docs", n)})
		}
	} else if n, ok := asInt(rules["km_documents"]); ok {
		items = append(items, item{2, fmt.Sprintf("%d knowledge docs", n)})
	}
	if n, ok := asInt(rules["max_monthly_call_minutes"]); ok {
		if n >= 10_000_000 {
			items = append(items, item{3, "Unlimited platform voice minutes (BYOK)"})
		} else {
			items = append(items, item{3, fmt.Sprintf("%d call minutes/mo", n)})
		}
	}
	if n, ok := asInt(rules["max_concurrent_calls"]); ok {
		items = append(items, item{4, fmt.Sprintf("%d concurrent voice sessions", n)})
	}
	// Avoid double-listing storage when already shown in the Unlimited KM line.
	if !storageInKMLine {
		if n, ok := asInt(rules["max_storage_bytes"]); ok && n > 0 {
			gb := n / (1 << 30)
			if gb < 1 {
				gb = 1
			}
			items = append(items, item{5, fmt.Sprintf("%d GB knowledge storage", gb)})
		} else if n, ok := asInt(rules["max_storage_gb"]); ok {
			items = append(items, item{5, fmt.Sprintf("%d GB storage", n)})
		} else if n, ok := asInt(rules["storage_gb"]); ok {
			items = append(items, item{5, fmt.Sprintf("%d GB storage", n)})
		}
	}
	if n, ok := asInt(rules["max_ai_employees"]); ok && n >= 1_000_000 {
		// replace avatar line with unlimited
		for i := range items {
			if items[i].order == 1 {
				items[i].text = "Unlimited AI avatars"
			}
		}
	}
	if b, ok := asBool(rules["voice_enabled"]); ok && b {
		items = append(items, item{6, "Voice enabled · BYOK AI"})
	}
	if b, ok := asBool(rules["rag_enabled"]); ok && b {
		items = append(items, item{7, "Knowledge / RAG enabled"})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].order < items[j].order })
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.text)
	}
	return out
}

func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case float32:
		return int(n), true
	default:
		return 0, false
	}
}

func asBool(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}
