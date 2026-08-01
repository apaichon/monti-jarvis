package quota

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	queueUpdateStatus   = "queue_status"
	queueUpdateAdmitted = "queue_admitted"
	queueUpdateTimeout  = "queue_timeout"

	busyStatusAvailable = "available"
	busyStatusBusy      = "busy"
	busyStatusQueued    = "queued"
	busyStatusAdmitted  = "admitted"
	busyStatusLive      = "live"
	busyStatusTimeout   = "timeout"
)

// WaitForQueuedConcurrent reserves a concurrent call slot immediately or keeps
// the caller in a tenant FIFO queue until capacity opens or the wait expires.
func (s *Service) WaitForQueuedConcurrent(ctx context.Context, tenantID, admissionID string, notify func(QueueUpdate) error) (*QueuedAdmission, error) {
	noop := func() {}
	if s == nil || !s.enabled || !s.callQueueEnabled {
		release, err := s.AcquireConcurrent(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		snap, _ := s.ConcurrentQueueSnapshot(ctx, tenantID)
		snap.BusyStatus = busyStatusAdmitted
		return &QueuedAdmission{AdmissionID: strings.TrimSpace(admissionID), Release: release, Snapshot: snap}, nil
	}
	if s.rdb == nil || strings.TrimSpace(tenantID) == "" {
		return &QueuedAdmission{AdmissionID: strings.TrimSpace(admissionID), Release: noop}, nil
	}

	release, err := s.AcquireConcurrent(ctx, tenantID)
	if err == nil {
		snap, _ := s.ConcurrentQueueSnapshot(ctx, tenantID)
		snap.BusyStatus = busyStatusAdmitted
		return &QueuedAdmission{AdmissionID: strings.TrimSpace(admissionID), Release: release, Snapshot: snap}, nil
	}
	if !isConcurrentLimit(err) {
		return nil, err
	}

	admissionID = strings.TrimSpace(admissionID)
	if admissionID == "" {
		admissionID = newAdmissionID()
	}
	inserted, err := s.enqueueCaller(ctx, tenantID, admissionID)
	if err != nil {
		return nil, err
	}
	queued := true
	defer func() {
		if queued {
			_ = s.removeQueuedCaller(context.Background(), tenantID, admissionID)
		}
	}()
	if inserted && s.callQueueMaxPerTenant > 0 {
		count, err := s.rdb.ZCard(ctx, s.callQueueKey(tenantID)).Result()
		if err != nil {
			if e := s.onRedisErr("WaitForQueuedConcurrent ZCard", err); e != nil {
				return nil, e
			}
		}
		if int(count) > s.callQueueMaxPerTenant {
			queued = false
			_ = s.removeQueuedCaller(ctx, tenantID, admissionID)
			return nil, QueueFull(s.callQueueMaxPerTenant, int(count))
		}
	}

	maxWait := s.queueMaxWait()
	ticker := time.NewTicker(s.queueRefresh())
	defer ticker.Stop()
	timer := time.NewTimer(maxWait)
	defer timer.Stop()

	for {
		admitted, err := s.tryPromoteQueuedCaller(ctx, tenantID, admissionID)
		if err != nil {
			return nil, err
		}
		if admitted != nil {
			queued = false
			if notify != nil {
				if err := notify(queueUpdate(queueUpdateAdmitted, admissionID, 0, *admitted)); err != nil {
					if admitted.Release != nil {
						admitted.Release()
					}
					return nil, err
				}
			}
			return admitted, nil
		}

		update, err := s.currentQueueUpdate(ctx, tenantID, admissionID)
		if err != nil {
			return nil, err
		}
		if update.Position <= 0 {
			return nil, QueueTimeout(int(maxWait.Seconds()), int(maxWait.Seconds()))
		}
		if notify != nil {
			if err := notify(update); err != nil {
				return nil, err
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			queued = false
			_ = s.removeQueuedCaller(ctx, tenantID, admissionID)
			s.incrementQueueTimeouts(ctx, tenantID, 1)
			timeout := update.Snapshot
			timeout.BusyStatus = busyStatusTimeout
			if notify != nil {
				_ = notify(QueueUpdate{
					Type:        queueUpdateTimeout,
					AdmissionID: admissionID,
					Snapshot:    timeout,
					Message:     "Call queue wait timed out. Please try again.",
				})
			}
			return nil, QueueTimeout(int(maxWait.Seconds()), int(maxWait.Seconds()))
		case <-ticker.C:
		}
	}
}

// ConcurrentQueueSnapshot returns a tenant-scoped active/queued/total capacity view.
func (s *Service) ConcurrentQueueSnapshot(ctx context.Context, tenantID string) (ConcurrentQueueSnapshot, error) {
	out := ConcurrentQueueSnapshot{
		QueueEnabled: s != nil && s.enabled && s.callQueueEnabled,
		BusyStatus:   busyStatusAvailable,
	}
	if s == nil {
		return out, nil
	}
	limits, err := s.limitsOrNil(ctx, tenantID)
	if err != nil {
		if errors.Is(err, ErrNoEntitlement) {
			return out, nil
		}
		return out, err
	}
	if limits != nil {
		out.MaxConcurrentCalls = limits.MaxConcurrentCalls
	}
	if s.rdb == nil || strings.TrimSpace(tenantID) == "" {
		return out, nil
	}
	_ = s.cleanupExpiredQueue(ctx, tenantID)
	active, err := s.getInt(ctx, s.concurrentKey(tenantID))
	if err != nil {
		if e := s.onRedisErr("ConcurrentQueueSnapshot active", err); e != nil {
			return out, e
		}
	}
	if active < 0 {
		active = 0
	}
	queued64, err := s.rdb.ZCard(ctx, s.callQueueKey(tenantID)).Result()
	if err != nil {
		if e := s.onRedisErr("ConcurrentQueueSnapshot queued", err); e != nil {
			return out, e
		}
	}
	out.ActiveCalls = active
	out.QueuedCallers = int(queued64)
	out.TotalCalls = out.ActiveCalls + out.QueuedCallers
	if out.QueuedCallers > 0 {
		out.BusyStatus = busyStatusQueued
	} else if out.MaxConcurrentCalls > 0 && out.ActiveCalls >= out.MaxConcurrentCalls {
		out.BusyStatus = busyStatusBusy
	} else if out.ActiveCalls > 0 {
		out.BusyStatus = busyStatusLive
	}
	out.OldestWaitSeconds = s.oldestQueueWaitSeconds(ctx, tenantID)
	out.RecentTimeouts24h, _ = s.getInt(ctx, s.callQueueTimeoutsKey(tenantID))
	return out, nil
}

func (s *Service) tryPromoteQueuedCaller(ctx context.Context, tenantID, admissionID string) (*QueuedAdmission, error) {
	if err := s.cleanupExpiredQueue(ctx, tenantID); err != nil {
		return nil, err
	}
	position, exists, err := s.queuedCallerPosition(ctx, tenantID, admissionID)
	if err != nil || !exists || position != 1 {
		return nil, err
	}
	locked, unlock, err := s.acquirePromotionLock(ctx, tenantID)
	if err != nil || !locked {
		return nil, err
	}
	defer unlock()

	position, exists, err = s.queuedCallerPosition(ctx, tenantID, admissionID)
	if err != nil || !exists || position != 1 {
		return nil, err
	}
	release, err := s.AcquireConcurrent(ctx, tenantID)
	if err != nil {
		if isConcurrentLimit(err) {
			return nil, nil
		}
		return nil, err
	}
	if err := s.removeQueuedCaller(ctx, tenantID, admissionID); err != nil {
		release()
		return nil, err
	}
	snap, err := s.ConcurrentQueueSnapshot(ctx, tenantID)
	if err != nil {
		release()
		return nil, err
	}
	snap.BusyStatus = busyStatusAdmitted
	return &QueuedAdmission{AdmissionID: admissionID, Release: release, Snapshot: snap}, nil
}

func (s *Service) currentQueueUpdate(ctx context.Context, tenantID, admissionID string) (QueueUpdate, error) {
	if err := s.cleanupExpiredQueue(ctx, tenantID); err != nil {
		return QueueUpdate{}, err
	}
	position, exists, err := s.queuedCallerPosition(ctx, tenantID, admissionID)
	if err != nil || !exists {
		return QueueUpdate{}, err
	}
	snap, err := s.ConcurrentQueueSnapshot(ctx, tenantID)
	if err != nil {
		return QueueUpdate{}, err
	}
	snap.BusyStatus = busyStatusQueued
	update := QueueUpdate{
		Type:                 queueUpdateStatus,
		AdmissionID:          admissionID,
		Position:             position,
		EstimatedWaitSeconds: s.estimatedQueueWaitSeconds(ctx, tenantID, admissionID, position),
		Snapshot:             snap,
	}
	update.Message = fmt.Sprintf("All agents are busy. You are #%d in line.", position)
	return update, nil
}

func queueUpdate(kind, admissionID string, position int, admitted QueuedAdmission) QueueUpdate {
	return QueueUpdate{
		Type:        kind,
		AdmissionID: admissionID,
		Position:    position,
		Snapshot:    admitted.Snapshot,
		Message:     "A call slot is available. Connecting now.",
	}
}

func (s *Service) enqueueCaller(ctx context.Context, tenantID, admissionID string) (bool, error) {
	key := s.callQueueKey(tenantID)
	now := s.queueNow()
	added, err := s.rdb.ZAddNX(ctx, key, redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: admissionID,
	}).Result()
	if err != nil {
		if e := s.onRedisErr("enqueueCaller", err); e != nil {
			return false, e
		}
		return false, nil
	}
	ttl := s.queueMaxWait() + time.Minute
	_ = s.rdb.HSet(ctx, s.callQueueEntryKey(tenantID, admissionID), map[string]any{
		"admission_id":  admissionID,
		"tenant_id":     tenantID,
		"status":        busyStatusQueued,
		"created_at_ms": now.UnixMilli(),
		"expires_at_ms": now.Add(s.queueMaxWait()).UnixMilli(),
		"updated_at_ms": now.UnixMilli(),
	}).Err()
	_ = s.rdb.Expire(ctx, s.callQueueEntryKey(tenantID, admissionID), ttl).Err()
	_ = s.rdb.Expire(ctx, key, s.queueMaxWait()+s.concurrentTTL).Err()
	return added > 0, nil
}

func (s *Service) removeQueuedCaller(ctx context.Context, tenantID, admissionID string) error {
	if s == nil || s.rdb == nil || strings.TrimSpace(admissionID) == "" {
		return nil
	}
	if err := s.rdb.ZRem(ctx, s.callQueueKey(tenantID), admissionID).Err(); err != nil {
		return s.onRedisErr("removeQueuedCaller", err)
	}
	if err := s.rdb.Del(ctx, s.callQueueEntryKey(tenantID, admissionID)).Err(); err != nil {
		return s.onRedisErr("removeQueuedCaller entry", err)
	}
	return nil
}

func (s *Service) queuedCallerPosition(ctx context.Context, tenantID, admissionID string) (int, bool, error) {
	rank, err := s.rdb.ZRank(ctx, s.callQueueKey(tenantID), admissionID).Result()
	if errors.Is(err, redis.Nil) {
		return 0, false, nil
	}
	if err != nil {
		if e := s.onRedisErr("queuedCallerPosition", err); e != nil {
			return 0, false, e
		}
		return 0, false, nil
	}
	return int(rank) + 1, true, nil
}

func (s *Service) cleanupExpiredQueue(ctx context.Context, tenantID string) error {
	if s == nil || s.rdb == nil || strings.TrimSpace(tenantID) == "" {
		return nil
	}
	cutoff := s.queueNow().Add(-s.queueMaxWait()).UnixMilli()
	key := s.callQueueKey(tenantID)
	expired, err := s.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: "-inf",
		Max: strconv.FormatInt(cutoff, 10),
	}).Result()
	if err != nil {
		if e := s.onRedisErr("cleanupExpiredQueue range", err); e != nil {
			return e
		}
		return nil
	}
	if len(expired) == 0 {
		return nil
	}
	members := make([]interface{}, 0, len(expired))
	for _, id := range expired {
		members = append(members, id)
	}
	removed, err := s.rdb.ZRem(ctx, key, members...).Result()
	if err != nil {
		if e := s.onRedisErr("cleanupExpiredQueue remove", err); e != nil {
			return e
		}
		return nil
	}
	if removed > 0 {
		for _, id := range expired {
			_ = s.rdb.Del(ctx, s.callQueueEntryKey(tenantID, id)).Err()
		}
		s.incrementQueueTimeouts(ctx, tenantID, removed)
	}
	return nil
}

func (s *Service) oldestQueueWaitSeconds(ctx context.Context, tenantID string) int {
	items, err := s.rdb.ZRangeWithScores(ctx, s.callQueueKey(tenantID), 0, 0).Result()
	if err != nil || len(items) == 0 {
		return 0
	}
	wait := s.queueNow().Sub(time.UnixMilli(int64(items[0].Score)))
	if wait < 0 {
		return 0
	}
	return int(wait.Seconds())
}

func (s *Service) estimatedQueueWaitSeconds(ctx context.Context, tenantID, admissionID string, position int) int {
	if position <= 0 {
		return 0
	}
	score, err := s.rdb.ZScore(ctx, s.callQueueKey(tenantID), admissionID).Result()
	if err != nil {
		return position * 30
	}
	remaining := s.queueMaxWait() - s.queueNow().Sub(time.UnixMilli(int64(score)))
	if remaining <= 0 {
		return 0
	}
	estimate := position * 30
	if estimate > int(remaining.Seconds()) {
		return int(remaining.Seconds())
	}
	return estimate
}

func (s *Service) acquirePromotionLock(ctx context.Context, tenantID string) (bool, func(), error) {
	key := s.callQueuePromotionLockKey(tenantID)
	token := newAdmissionID()
	ok, err := s.rdb.SetNX(ctx, key, token, s.queueLockTTL()).Result()
	if err != nil {
		if e := s.onRedisErr("acquirePromotionLock", err); e != nil {
			return false, func() {}, e
		}
		return false, func() {}, nil
	}
	return ok, func() {
		rctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if current, err := s.rdb.Get(rctx, key).Result(); err == nil && current == token {
			_ = s.rdb.Del(rctx, key).Err()
		}
	}, nil
}

func (s *Service) incrementQueueTimeouts(ctx context.Context, tenantID string, count int64) {
	if s == nil || s.rdb == nil || count <= 0 {
		return
	}
	key := s.callQueueTimeoutsKey(tenantID)
	if n, err := s.rdb.IncrBy(ctx, key, count).Result(); err == nil && n == count {
		_ = s.rdb.Expire(ctx, key, 24*time.Hour).Err()
	}
}

func (s *Service) callQueueKey(tenantID string) string {
	return s.prefix + "callq:" + tenantID + ":voice"
}

func (s *Service) callQueueEntryKey(tenantID, admissionID string) string {
	return s.prefix + "callq:" + tenantID + ":entry:" + admissionID
}

func (s *Service) callQueueTimeoutsKey(tenantID string) string {
	return s.prefix + "callq:" + tenantID + ":recent_timeouts"
}

func (s *Service) callQueuePromotionLockKey(tenantID string) string {
	return s.prefix + "callq:" + tenantID + ":promotion_lock"
}

func (s *Service) queueMaxWait() time.Duration {
	if s == nil || s.callQueueMaxWait <= 0 {
		return 120 * time.Second
	}
	return s.callQueueMaxWait
}

func (s *Service) queueRefresh() time.Duration {
	if s == nil || s.callQueuePositionRefresh <= 0 {
		return 2 * time.Second
	}
	return s.callQueuePositionRefresh
}

func (s *Service) queueLockTTL() time.Duration {
	if s == nil || s.callQueuePromotionLockTTL <= 0 {
		return 10 * time.Second
	}
	return s.callQueuePromotionLockTTL
}

func (s *Service) queueNow() time.Time {
	if s != nil && s.now != nil {
		return s.now()
	}
	return time.Now()
}

func isConcurrentLimit(err error) bool {
	var qe *Error
	return errors.As(err, &qe) && qe.Code == "quota_exceeded" && qe.Dimension == DimMaxConcurrentCalls
}

func newAdmissionID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err == nil {
		return "adm_" + hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("adm_%d", time.Now().UnixNano())
}
