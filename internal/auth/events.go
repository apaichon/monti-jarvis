package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/natsbus"
)

type EventPublisher struct {
	bus *natsbus.Bus
	ch  *clickhouse.Client
	on  bool
}

func NewEventPublisher(bus *natsbus.Bus, ch *clickhouse.Client, enabled bool) *EventPublisher {
	return &EventPublisher{bus: bus, ch: ch, on: enabled}
}

func (p *EventPublisher) Enabled() bool {
	return p != nil && p.on && p.bus != nil && p.bus.Enabled()
}

func (p *EventPublisher) Publish(ctx context.Context, subject string, user CachedUser, meta RequestMeta, extra map[string]any) {
	if p == nil || !p.on {
		return
	}
	event := natsbus.AuthEvent{
		EventID:   newEventID(),
		Event:     subject,
		TenantID:  user.TenantID,
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		IP:        meta.IP,
		UserAgent: meta.UserAgent,
		At:        time.Now().UTC(),
		Meta:      extra,
	}
	if p.bus != nil && p.bus.Enabled() {
		_ = p.bus.PublishAuthEvent(ctx, event)
	}
	if p.ch != nil && p.ch.Enabled() {
		_ = p.ch.InsertAuthEvent(ctx, clickhouse.AuthEventRow{
			EventID:   event.EventID,
			Event:     event.Event,
			TenantID:  event.TenantID,
			UserID:    event.UserID,
			Email:     event.Email,
			Role:      event.Role,
			IP:        event.IP,
			UserAgent: event.UserAgent,
			CreatedAt: event.At,
		})
	}
}

func (p *EventPublisher) PublishFailure(ctx context.Context, email string, meta RequestMeta) {
	if p == nil || !p.on {
		return
	}
	event := natsbus.AuthEvent{
		EventID:   newEventID(),
		Event:     "auth.login.failed",
		Email:     email,
		IP:        meta.IP,
		UserAgent: meta.UserAgent,
		At:        time.Now().UTC(),
	}
	if p.bus != nil && p.bus.Enabled() {
		_ = p.bus.PublishAuthEvent(ctx, event)
	}
	if p.ch != nil && p.ch.Enabled() {
		_ = p.ch.InsertAuthEvent(ctx, clickhouse.AuthEventRow{
			EventID:   event.EventID,
			Event:     event.Event,
			Email:     event.Email,
			IP:        event.IP,
			UserAgent: event.UserAgent,
			CreatedAt: event.At,
		})
	}
}

func newEventID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "evt_" + time.Now().UTC().Format("20060102150405")
	}
	return "evt_" + hex.EncodeToString(b[:])
}