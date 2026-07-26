package store

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/libra/monti-jarvis/internal/env"
)

func TestCapabilityPoolsPreferDedicatedConnections(t *testing.T) {
	writer := &pgxpool.Pool{}
	kmRead := &pgxpool.Pool{}
	ticketWrite := &pgxpool.Pool{}
	s := &Store{pg: writer, pgKMRead: kmRead, pgTicketWrite: ticketWrite}
	if got := s.KMReadDB(); got != kmRead {
		t.Fatalf("KM read pool = %p, want dedicated pool %p", got, kmRead)
	}
	if got := s.TicketWriteDB(); got != ticketWrite {
		t.Fatalf("ticket write pool = %p, want dedicated pool %p", got, ticketWrite)
	}
}

func TestCapabilityPoolsFallbackForDevelopment(t *testing.T) {
	writer := &pgxpool.Pool{}
	s := &Store{pg: writer}
	if got := s.KMReadDB(); got != writer {
		t.Fatalf("KM read fallback = %p, want writer %p", got, writer)
	}
	if got := s.TicketWriteDB(); got != writer {
		t.Fatalf("ticket write fallback = %p, want writer %p", got, writer)
	}
}

func TestValidateCapabilityPoolsAllowsDevelopmentFallback(t *testing.T) {
	s := &Store{cfg: testCapabilityConfig("dev"), pg: &pgxpool.Pool{}}
	if err := s.ValidateCapabilityPools(); err != nil {
		t.Fatalf("development fallback rejected: %v", err)
	}
}

func TestValidateCapabilityPoolsRejectsProductionFallback(t *testing.T) {
	s := &Store{cfg: testCapabilityConfig("production"), pg: &pgxpool.Pool{}}
	if err := s.ValidateCapabilityPools(); err == nil {
		t.Fatal("expected missing capability pool error")
	}
}

func testCapabilityConfig(appEnv string) env.Config {
	return env.Config{AppEnv: appEnv}
}
