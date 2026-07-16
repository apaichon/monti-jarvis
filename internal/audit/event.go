package audit

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Event is the redacted, immutable audit envelope persisted by the writer.
type Event struct {
	EventID      string         `json:"event_id"`
	OccurredAt   time.Time      `json:"occurred_at"`
	TenantID     string         `json:"tenant_id"`
	ActorID      string         `json:"actor_id"`
	ActorType    string         `json:"actor_type"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	RequestID    string         `json:"request_id"`
	Source       string         `json:"source"`
	Outcome      string         `json:"outcome"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type Actor struct {
	ID       string
	Type     string
	TenantID string
}

func (e Event) Validate() error {
	if strings.TrimSpace(e.EventID) == "" {
		return fmt.Errorf("event_id is required")
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("occurred_at is required")
	}
	for name, value := range map[string]string{
		"action":        e.Action,
		"resource_type": e.ResourceType,
		"request_id":    e.RequestID,
		"source":        e.Source,
		"outcome":       e.Outcome,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

func NewEvent(actor Actor, action, resourceType, resourceID, requestID, source, outcome string, metadata map[string]any) Event {
	if strings.TrimSpace(actor.ID) == "" {
		actor.ID = "system"
	}
	if strings.TrimSpace(actor.Type) == "" {
		actor.Type = "system"
	}
	return Event{
		EventID:      newID(),
		OccurredAt:   time.Now().UTC(),
		TenantID:     strings.TrimSpace(actor.TenantID),
		ActorID:      strings.TrimSpace(actor.ID),
		ActorType:    strings.TrimSpace(actor.Type),
		Action:       strings.TrimSpace(action),
		ResourceType: strings.TrimSpace(resourceType),
		ResourceID:   strings.TrimSpace(resourceID),
		RequestID:    strings.TrimSpace(requestID),
		Source:       strings.TrimSpace(source),
		Outcome:      strings.TrimSpace(outcome),
		Metadata:     metadata,
	}
}

func marshalEvent(e Event) ([]byte, error) {
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	if e.EventID == "" {
		e.EventID = newID()
	}
	if e.ActorID == "" {
		e.ActorID = "system"
	}
	if e.ActorType == "" {
		e.ActorType = "system"
	}
	if err := e.Validate(); err != nil {
		return nil, err
	}
	e.Metadata = sanitizeMetadata(e.Metadata)
	return json.Marshal(e)
}

func sanitizeMetadata(input map[string]any) map[string]any {
	if len(input) == 0 {
		return nil
	}
	allowed := map[string]struct{}{
		"method": {}, "path": {}, "status": {}, "duration_ms": {}, "avatar_id": {},
		"resource_id": {}, "changed_fields": {}, "reason": {}, "operation": {}, "error_code": {},
	}
	output := make(map[string]any)
	for key, value := range input {
		if _, ok := allowed[key]; !ok {
			continue
		}
		encoded, err := json.Marshal(value)
		if err != nil || len(encoded) > 1024 {
			continue
		}
		output[key] = value
	}
	if len(output) == 0 {
		return nil
	}
	return output
}

func newID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return "evt_" + hex.EncodeToString(b[:])
}
