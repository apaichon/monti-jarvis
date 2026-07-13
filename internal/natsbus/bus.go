package natsbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Bus struct {
	conn *nats.Conn
}

type CallEvent struct {
	Event     string    `json:"event"`
	SessionID string    `json:"session_id"`
	TenantID  string    `json:"tenant_id"`
	RoomName  string    `json:"room_name,omitempty"`
	Role      string    `json:"role,omitempty"`
	Content   string    `json:"content,omitempty"`
	At        time.Time `json:"at"`
}

func Connect(url string) (*Bus, error) {
	if url == "" {
		return nil, fmt.Errorf("nats url is empty")
	}
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &Bus{conn: nc}, nil
}

func (b *Bus) Close() {
	if b != nil && b.conn != nil {
		b.conn.Close()
	}
}

func (b *Bus) Enabled() bool {
	return b != nil && b.conn != nil && b.conn.IsConnected()
}

func (b *Bus) PublishCallEvent(ctx context.Context, subject string, event CallEvent) error {
	if !b.Enabled() {
		return fmt.Errorf("nats is not connected")
	}
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return b.conn.Publish(subject, payload)
}

// PublishJSON publishes a bounded integration event when NATS is available.
func (b *Bus) PublishJSON(subject string, value any) error {
	if !b.Enabled() {
		return fmt.Errorf("nats is not connected")
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return b.conn.Publish(subject, payload)
}

func Subject(event string) string {
	return event
}
