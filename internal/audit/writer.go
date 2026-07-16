package audit

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Sink interface {
	InsertAuditEvents(context.Context, []Event) error
}

type Config struct {
	Mode          string
	Dir           string
	FlushInterval time.Duration
	Retention     time.Duration
	BatchSize     int
	QueueSize     int
	RetryBackoff  time.Duration
}

type Health struct {
	Mode                   string
	QueueDepth             int
	LastSuccessfulTransfer *time.Time
	OldestPendingFileAge   time.Duration
	PendingFiles           int
	FailedFiles            int
}

type Writer struct {
	cfg  Config
	sink Sink

	mu         sync.Mutex
	file       *os.File
	activePath string
	lastOK     time.Time
	failed     map[string]time.Time
	started    bool
	cancel     context.CancelFunc
	done       chan struct{}
}

func New(cfg Config, sink Sink) (*Writer, error) {
	cfg.Mode = strings.ToLower(strings.TrimSpace(cfg.Mode))
	if cfg.Mode == "" {
		cfg.Mode = "spool"
	}
	if cfg.Mode != "spool" && cfg.Mode != "clickhouse" {
		return nil, fmt.Errorf("invalid audit log mode %q", cfg.Mode)
	}
	if cfg.Dir == "" {
		cfg.Dir = "./var/audit"
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 5 * time.Second
	}
	if cfg.Retention <= 0 {
		cfg.Retention = time.Hour
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 10000
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = time.Second
	}
	if err := os.MkdirAll(cfg.Dir, 0o750); err != nil {
		return nil, fmt.Errorf("audit spool directory: %w", err)
	}
	return &Writer{cfg: cfg, sink: sink, failed: make(map[string]time.Time)}, nil
}

func (w *Writer) Start(ctx context.Context) {
	w.mu.Lock()
	if w.started {
		w.mu.Unlock()
		return
	}
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.done = make(chan struct{})
	w.started = true
	w.mu.Unlock()
	go w.run(workerCtx)
}

func (w *Writer) Close(ctx context.Context) error {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
	}
	done := w.done
	w.mu.Unlock()
	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		_ = w.file.Sync()
		err := w.file.Close()
		w.file = nil
		w.activePath = ""
		return err
	}
	return nil
}

func (w *Writer) Enqueue(event Event) error {
	return w.EnqueueContext(context.Background(), event)
}

func (w *Writer) EnqueueContext(ctx context.Context, event Event) error {
	data, err := marshalEvent(event)
	if err != nil {
		return err
	}
	if len(data) > 256*1024 {
		return fmt.Errorf("audit event exceeds 256KB")
	}
	if w.cfg.Mode == "clickhouse" && w.sink != nil {
		directCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := w.sink.InsertAuditEvents(directCtx, []Event{event})
		cancel()
		if err == nil {
			w.mu.Lock()
			w.lastOK = time.Now().UTC()
			w.mu.Unlock()
			return nil
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.appendLocked(data); err != nil {
		return err
	}
	return nil
}

func (w *Writer) appendLocked(data []byte) error {
	if err := w.ensureActiveLocked(); err != nil {
		return err
	}
	if _, err := w.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}
	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("sync audit event: %w", err)
	}
	return nil
}

func (w *Writer) Health() Health {
	w.mu.Lock()
	defer w.mu.Unlock()
	h := Health{Mode: w.cfg.Mode}
	if !w.lastOK.IsZero() {
		last := w.lastOK
		h.LastSuccessfulTransfer = &last
	}
	entries, err := os.ReadDir(w.cfg.Dir)
	if err != nil {
		return h
	}
	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") || entry.Name() == filepath.Base(w.activePath) {
			continue
		}
		h.PendingFiles++
		if info, err := entry.Info(); err == nil {
			age := now.Sub(info.ModTime())
			if age > h.OldestPendingFileAge {
				h.OldestPendingFileAge = age
			}
		}
		if _, ok := w.failed[filepath.Join(w.cfg.Dir, entry.Name())]; ok {
			h.FailedFiles++
		}
	}
	return h
}

func (w *Writer) run(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.FlushInterval)
	defer ticker.Stop()
	defer func() {
		w.mu.Lock()
		_ = w.rotateLocked()
		w.mu.Unlock()
		_ = w.flushOnce(context.Background())
		w.mu.Lock()
		if w.done != nil {
			close(w.done)
		}
		w.mu.Unlock()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = w.flushOnce(ctx)
		}
	}
}

func (w *Writer) flushOnce(ctx context.Context) error {
	w.mu.Lock()
	if err := w.rotateLocked(); err != nil {
		w.mu.Unlock()
		return err
	}
	w.mu.Unlock()
	if w.sink != nil {
		for _, path := range w.closedFiles() {
			if err := w.transferFile(ctx, path); err != nil {
				w.mu.Lock()
				w.failed[path] = time.Now()
				w.mu.Unlock()
			}
		}
	}
	w.cleanupAcknowledged()
	return nil
}

func (w *Writer) ensureActiveLocked() error {
	if w.file != nil {
		return nil
	}
	stamp := time.Now().UTC().Format("20060102-15-04-05")
	path := filepath.Join(w.cfg.Dir, "audit_log_"+stamp+".jsonl")
	if _, err := os.Stat(path + ".uploaded"); err == nil {
		path = filepath.Join(w.cfg.Dir, "audit_log_"+stamp+"-"+newID()[4:]+".jsonl")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	w.file = f
	w.activePath = path
	return nil
}

func (w *Writer) rotateLocked() error {
	if w.file == nil {
		return nil
	}
	if err := w.file.Sync(); err != nil {
		return err
	}
	err := w.file.Close()
	w.file = nil
	w.activePath = ""
	return err
}

func (w *Writer) closedFiles() []string {
	entries, err := os.ReadDir(w.cfg.Dir)
	if err != nil {
		return nil
	}
	w.mu.Lock()
	activePath := w.activePath
	w.mu.Unlock()
	var paths []string
	for _, entry := range entries {
		path := filepath.Join(w.cfg.Dir, entry.Name())
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "audit_log_") || !strings.HasSuffix(entry.Name(), ".jsonl") || path == activePath {
			continue
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func (w *Writer) transferFile(ctx context.Context, path string) error {
	if _, err := os.Stat(path + ".uploaded"); err == nil {
		return nil
	}
	events, size, checksum, err := readEvents(path)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return writeAck(path, 0, size, checksum)
	}
	for start := 0; start < len(events); start += w.cfg.BatchSize {
		end := start + w.cfg.BatchSize
		if end > len(events) {
			end = len(events)
		}
		if err := w.sink.InsertAuditEvents(ctx, events[start:end]); err != nil {
			return err
		}
	}
	if err := writeAck(path, len(events), size, checksum); err != nil {
		return err
	}
	w.mu.Lock()
	w.lastOK = time.Now().UTC()
	delete(w.failed, path)
	w.mu.Unlock()
	return nil
}

func readEvents(path string) ([]Event, int64, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, 0, "", err
	}
	hash := sha256.New()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 512*1024)
	var events []Event
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, info.Size(), "", fmt.Errorf("decode %s: %w", filepath.Base(path), err)
		}
		events = append(events, event)
		_, _ = hash.Write(append(append([]byte(nil), line...), '\n'))
	}
	if err := scanner.Err(); err != nil {
		return nil, info.Size(), "", err
	}
	return events, info.Size(), hex.EncodeToString(hash.Sum(nil)), nil
}

func writeAck(path string, count int, size int64, checksum string) error {
	ack := map[string]any{"event_count": count, "byte_count": size, "checksum": checksum, "acknowledged_at": time.Now().UTC()}
	b, err := json.Marshal(ack)
	if err != nil {
		return err
	}
	tmp := path + ".uploaded.tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o640); err != nil {
		return err
	}
	return os.Rename(tmp, path+".uploaded")
}

func (w *Writer) cleanupAcknowledged() {
	entries, err := os.ReadDir(w.cfg.Dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-w.cfg.Retention)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".uploaded") {
			continue
		}
		path := filepath.Join(w.cfg.Dir, strings.TrimSuffix(entry.Name(), ".uploaded"))
		info, err := os.Stat(path)
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}
		if err := os.Remove(path); err == nil {
			_ = os.Remove(filepath.Join(w.cfg.Dir, entry.Name()))
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// Middleware records mutating API outcomes without capturing request bodies.
func (w *Writer) Middleware(next http.Handler, resolve func(*http.Request) Actor, defaultTenant string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions ||
			strings.HasPrefix(r.URL.Path, "/api/platform/audit-logs") || strings.HasPrefix(r.URL.Path, "/ws/") {
			next.ServeHTTP(rw, r)
			return
		}
		started := time.Now()
		wrapped := &responseWriter{ResponseWriter: rw}
		next.ServeHTTP(wrapped, r)
		actor := Actor{ID: "anonymous", Type: "anonymous", TenantID: strings.TrimSpace(r.Header.Get("X-Tenant-Id"))}
		if resolve != nil {
			actor = resolve(r)
		}
		if actor.TenantID == "" {
			actor.TenantID = strings.TrimSpace(defaultTenant)
		}
		status := wrapped.status
		if status == 0 {
			status = http.StatusOK
		}
		outcome := "success"
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			outcome = "denied"
		} else if status >= 400 {
			outcome = "failure"
		}
		resourceType, resourceID := resourceFromPath(r.URL.Path)
		metadata := map[string]any{"method": r.Method, "path": r.URL.Path, "status": status, "duration_ms": time.Since(started).Milliseconds()}
		event := NewEvent(actor, "http."+strings.ToLower(r.Method)+"."+resourceType, resourceType, resourceID, requestID(r), "http_api", outcome, metadata)
		_ = w.EnqueueContext(r.Context(), event)
	})
}

func requestID(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-Request-ID")); value != "" && len(value) <= 128 {
		return value
	}
	return "req_" + newID()[4:]
}

func resourceFromPath(path string) (string, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "api" || part == "platform" || part == "tenant" || part == "customer" || part == "public" {
			continue
		}
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	if len(filtered) == 0 {
		return "unknown", ""
	}
	resource := strings.TrimSuffix(filtered[0], "s")
	resourceID := ""
	if len(filtered) > 1 && !isAction(filtered[len(filtered)-1]) {
		resourceID = filtered[len(filtered)-1]
	}
	return resource, resourceID
}

func isAction(value string) bool {
	switch value {
	case "create", "update", "delete", "reset", "approve", "reject", "end", "rating", "login", "logout", "refresh", "test":
		return true
	default:
		return false
	}
}
