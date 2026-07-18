// Package metering normalizes provider usage into the reporting-only AI cost
// projection. It deliberately has no quota or payment side effects.
package metering

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/clickhouse"
)

const (
	StateObserved    = "observed"
	StateEstimated   = "estimated"
	StateUnavailable = "unavailable"
)

type UsageEvent struct {
	EventID              string
	TenantID             string
	CallID               string
	ConversationRecordID string
	Provider             string
	Model                string
	Modality             string
	InputTokens          uint64
	OutputTokens         uint64
	AudioSeconds         uint32
	UsageDate            time.Time
	SourceUpdatedAt      time.Time
}

type RateCard struct {
	Version                string
	Currency               string
	InputUnitPriceMicros   int64
	OutputUnitPriceMicros  int64
	AudioSecondPriceMicros int64
}

type Recorder struct {
	ch   *clickhouse.Client
	rate RateCard
	now  func() time.Time
}

func NewRecorder(ch *clickhouse.Client, rate RateCard) *Recorder {
	return &Recorder{ch: ch, rate: rate, now: time.Now}
}

func (r *Recorder) Record(ctx context.Context, event UsageEvent) error {
	if r == nil || r.ch == nil || !r.ch.Enabled() {
		return nil
	}
	if strings.TrimSpace(event.EventID) == "" {
		return nil
	}
	if event.UsageDate.IsZero() {
		event.UsageDate = r.now().UTC()
	}
	if event.SourceUpdatedAt.IsZero() {
		event.SourceUpdatedAt = r.now().UTC()
	}
	state, cost := r.price(event)
	return r.ch.UpsertAIUsageFact(ctx, clickhouse.AIUsageFact{
		FactID:           event.EventID,
		TenantID:         event.TenantID,
		CallID:           event.CallID,
		ConversationID:   event.ConversationRecordID,
		Provider:         event.Provider,
		Model:            event.Model,
		Modality:         event.Modality,
		MeasurementState: state,
		InputUnits:       event.InputTokens,
		OutputUnits:      event.OutputTokens,
		AudioSeconds:     event.AudioSeconds,
		RateVersion:      r.rate.Version,
		CostMicrounits:   cost,
		Currency:         r.rate.Currency,
		UsageDate:        event.UsageDate.UTC().Format("2006-01-02"),
		SourceUpdatedAt:  event.SourceUpdatedAt.UTC().Format("2006-01-02 15:04:05"),
		UpdatedAt:        r.now().UTC().Format("2006-01-02 15:04:05"),
	})
}

func (r *Recorder) price(event UsageEvent) (string, int64) {
	if r.rate.InputUnitPriceMicros <= 0 && r.rate.OutputUnitPriceMicros <= 0 && r.rate.AudioSecondPriceMicros <= 0 {
		return StateUnavailable, 0
	}
	if event.InputTokens > 0 || event.OutputTokens > 0 {
		if event.InputTokens > 0 && r.rate.InputUnitPriceMicros <= 0 {
			return StateUnavailable, 0
		}
		if event.OutputTokens > 0 && r.rate.OutputUnitPriceMicros <= 0 {
			return StateUnavailable, 0
		}
		cost := saturatingAdd(safeUnitCost(event.InputTokens, r.rate.InputUnitPriceMicros), safeUnitCost(event.OutputTokens, r.rate.OutputUnitPriceMicros))
		return StateObserved, cost
	}
	if event.AudioSeconds > 0 && r.rate.AudioSecondPriceMicros > 0 {
		return StateEstimated, safeDirectCost(uint64(event.AudioSeconds), r.rate.AudioSecondPriceMicros)
	}
	return StateUnavailable, 0
}

func saturatingAdd(left, right int64) int64 {
	if right > 0 && left > math.MaxInt64-right {
		return math.MaxInt64
	}
	return left + right
}

func safeDirectCost(units uint64, priceMicros int64) int64 {
	if units == 0 || priceMicros <= 0 || units > uint64(math.MaxInt64/priceMicros) {
		return 0
	}
	return int64(units * uint64(priceMicros))
}

func safeUnitCost(units uint64, priceMicros int64) int64 {
	if units == 0 || priceMicros <= 0 {
		return 0
	}
	if units > uint64(math.MaxInt64/priceMicros) {
		return math.MaxInt64
	}
	return int64(units * uint64(priceMicros) / 1_000_000)
}

// StableEventID creates an idempotency key without retaining source content.
func StableEventID(kind, tenantID, interactionID, sequence string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{kind, tenantID, interactionID, sequence}, "\x00")))
	return kind + "_" + hex.EncodeToString(sum[:16])
}
