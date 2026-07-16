package audit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

type fakeSink struct {
	events []Event
	err    error
}

func (f *fakeSink) InsertAuditEvents(_ context.Context, events []Event) error {
	if f.err != nil {
		return f.err
	}
	f.events = append(f.events, events...)
	return nil
}

func testWriter(t *testing.T, sink Sink) *Writer {
	t.Helper()
	w, err := New(Config{Dir: t.TempDir(), Retention: time.Hour, BatchSize: 2}, sink)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestWriterTransfersTimestampedJSONLAndAcknowledges(t *testing.T) {
	sink := &fakeSink{}
	w := testWriter(t, sink)
	event := NewEvent(Actor{ID: "user-1", Type: "platform_admin", TenantID: "tenant-1"}, "tenant.update", "tenant", "tenant-1", "req-1", "http_api", "success", map[string]any{"status": 200})
	if err := w.Enqueue(event); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(w.cfg.Dir, "audit_log_*.jsonl"))
	if err != nil || len(files) != 1 {
		t.Fatalf("expected one timestamped audit file, files=%v err=%v", files, err)
	}
	if ok, _ := regexp.MatchString(`audit_log_[0-9]{8}-[0-9]{2}-[0-9]{2}-[0-9]{2}(?:-[^.]*)?\.jsonl$`, filepath.Base(files[0])); !ok {
		t.Fatalf("unexpected audit filename %q", filepath.Base(files[0]))
	}
	if err := w.flushOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(sink.events) != 1 || sink.events[0].EventID != event.EventID {
		t.Fatalf("unexpected transferred events: %#v", sink.events)
	}
	if _, err := os.Stat(files[0] + ".uploaded"); err != nil {
		t.Fatalf("expected acknowledgement marker: %v", err)
	}
	if _, err := os.Stat(files[0]); err != nil {
		t.Fatalf("file must remain during retention: %v", err)
	}
}

func TestWriterRetainsFailedTransfer(t *testing.T) {
	w := testWriter(t, &fakeSink{err: errors.New("clickhouse unavailable")})
	if err := w.Enqueue(NewEvent(Actor{ID: "system", Type: "system"}, "system.rotate", "worker", "", "req-2", "worker", "success", nil)); err != nil {
		t.Fatal(err)
	}
	if err := w.flushOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(w.cfg.Dir, "audit_log_*.jsonl"))
	if len(files) != 1 {
		t.Fatalf("failed transfer must retain file, files=%v", files)
	}
	if _, err := os.Stat(files[0] + ".uploaded"); !os.IsNotExist(err) {
		t.Fatalf("failed transfer must not create acknowledgement marker")
	}
	if got := w.Health().FailedFiles; got != 1 {
		t.Fatalf("failed files=%d, want 1", got)
	}
}

func TestClickHouseModeUsesDirectSinkAndFallsBackToSpool(t *testing.T) {
	sink := &fakeSink{}
	w, err := New(Config{Mode: "clickhouse", Dir: t.TempDir(), Retention: time.Hour}, sink)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Enqueue(NewEvent(Actor{ID: "admin", Type: "platform_admin"}, "tenant.update", "tenant", "t1", "req-direct", "http_api", "success", nil)); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(w.cfg.Dir, "audit_log_*.jsonl"))
	if len(sink.events) != 1 || len(files) != 0 {
		t.Fatalf("direct mode should insert without spooling, events=%d files=%v", len(sink.events), files)
	}

	sink.err = errors.New("temporary failure")
	if err := w.Enqueue(NewEvent(Actor{ID: "admin", Type: "platform_admin"}, "tenant.update", "tenant", "t1", "req-fallback", "http_api", "success", nil)); err != nil {
		t.Fatal(err)
	}
	files, _ = filepath.Glob(filepath.Join(w.cfg.Dir, "audit_log_*.jsonl"))
	if len(files) != 1 {
		t.Fatalf("direct failure should fall back to local spool, files=%v", files)
	}
}

func TestMiddlewareCapturesMutationWithoutBody(t *testing.T) {
	w := testWriter(t, nil)
	h := w.Middleware(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusCreated)
	}), func(*http.Request) Actor {
		return Actor{ID: "admin-1", Type: "platform_admin", TenantID: "tenant-1"}
	}, "demo")
	req := httptest.NewRequest(http.MethodPost, "/api/platform/tenants/tenant-1/avatars", nil)
	req.Header.Set("X-Request-ID", "req-fixed")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	files, _ := filepath.Glob(filepath.Join(w.cfg.Dir, "audit_log_*.jsonl"))
	if len(files) != 1 {
		t.Fatalf("expected middleware event file, files=%v", files)
	}
	events, _, _, err := readEvents(files[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ActorID != "admin-1" || events[0].TenantID != "tenant-1" || events[0].RequestID != "req-fixed" {
		t.Fatalf("unexpected middleware event: %#v", events)
	}
	if events[0].Metadata["path"] != "/api/platform/tenants/tenant-1/avatars" || events[0].Metadata["body"] != nil {
		t.Fatalf("unexpected event path metadata: %#v", events[0].Metadata)
	}
}
