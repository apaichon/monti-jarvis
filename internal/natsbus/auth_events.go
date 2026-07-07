package natsbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const AuthStreamName = "MONTI_AUTH"

type AuthEvent struct {
	EventID   string         `json:"event_id"`
	Event     string         `json:"event"`
	TenantID  string         `json:"tenant_id,omitempty"`
	UserID    string         `json:"user_id,omitempty"`
	Email     string         `json:"email,omitempty"`
	Role      string         `json:"role,omitempty"`
	IP        string         `json:"ip,omitempty"`
	UserAgent string         `json:"user_agent,omitempty"`
	At        time.Time      `json:"at"`
	Meta      map[string]any `json:"meta,omitempty"`
}

func (b *Bus) EnsureAuthStream(ctx context.Context) error {
	if !b.Enabled() {
		return fmt.Errorf("nats is not connected")
	}
	js, err := jetstream.New(b.conn)
	if err != nil {
		return err
	}
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      AuthStreamName,
		Subjects:  []string{"auth.>"},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour,
	})
	return err
}

func (b *Bus) PublishAuthEvent(ctx context.Context, event AuthEvent) error {
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
	subject := event.Event
	if subject == "" {
		return fmt.Errorf("auth event subject is empty")
	}
	if err := b.conn.Publish(subject, payload); err != nil {
		return err
	}
	js, err := jetstream.New(b.conn)
	if err != nil {
		return err
	}
	_, err = js.Publish(ctx, subject, payload)
	return err
}

func (b *Bus) SubscribeAuthEvents(handler func(AuthEvent)) error {
	if !b.Enabled() {
		return fmt.Errorf("nats is not connected")
	}
	_, err := b.conn.Subscribe("auth.>", func(msg *nats.Msg) {
		var event AuthEvent
		if json.Unmarshal(msg.Data, &event) == nil {
			handler(event)
		}
	})
	return err
}