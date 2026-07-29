package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

const (
	BillingCycleScheduled       = "scheduled"
	BillingCyclePreviewed       = "previewed"
	BillingCyclePaymentPending  = "payment_pending"
	BillingCyclePaid            = "paid"
	BillingCycleDocumentsIssued = "documents_issued"
	BillingCycleSettled         = "settled"
	BillingCycleRetryWait       = "retry_wait"
	BillingCycleFailed          = "failed"
)

type BillingCycle struct {
	ID                  string           `json:"id"`
	SubscriptionID      string           `json:"subscription_id"`
	TenantID            string           `json:"tenant_id"`
	PackageID           string           `json:"package_id"`
	PeriodKey           string           `json:"period_key"`
	PeriodStart         time.Time        `json:"period_start"`
	PeriodEnd           time.Time        `json:"period_end"`
	Status              string           `json:"status"`
	CalculationSnapshot PriceCalculation `json:"calculation_snapshot"`
	OrderID             string           `json:"order_id,omitempty"`
	ReceiptID           string           `json:"receipt_id,omitempty"`
	TaxInvoiceID        string           `json:"tax_invoice_id,omitempty"`
	AttemptCount        int              `json:"attempt_count"`
	NextAttemptAt       *time.Time       `json:"next_attempt_at,omitempty"`
	LastErrorCode       string           `json:"last_error_code,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

func scanBillingCycle(row pgx.Row) (*BillingCycle, error) {
	var out BillingCycle
	var snapshot []byte
	var orderID, receiptID, taxInvoiceID *string
	err := row.Scan(
		&out.ID, &out.SubscriptionID, &out.TenantID, &out.PackageID,
		&out.PeriodKey, &out.PeriodStart, &out.PeriodEnd, &out.Status,
		&snapshot, &orderID, &receiptID, &taxInvoiceID, &out.AttemptCount,
		&out.NextAttemptAt, &out.LastErrorCode, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBillingCycleNotFound
	}
	if err != nil {
		return nil, err
	}
	if orderID != nil {
		out.OrderID = *orderID
	}
	if receiptID != nil {
		out.ReceiptID = *receiptID
	}
	if taxInvoiceID != nil {
		out.TaxInvoiceID = *taxInvoiceID
	}
	_ = json.Unmarshal(snapshot, &out.CalculationSnapshot)
	return &out, nil
}

const billingCycleSelect = `c.id, c.subscription_id, s.tenant_id, pv.package_id,
  c.period_key, c.period_start, c.period_end, c.status, c.calculation_snapshot,
  c.order_id, c.receipt_id, c.tax_invoice_id, c.attempt_count,
  c.next_attempt_at, c.last_error_code, c.created_at, c.updated_at`

func (s *Store) GetBillingCycle(ctx context.Context, id string) (*BillingCycle, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanBillingCycle(s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.billing_cycles c
JOIN %s.tenant_subscriptions s ON s.id = c.subscription_id
JOIN %s.package_versions pv ON pv.id = s.package_version_id
WHERE c.id = $1`, billingCycleSelect, schema, schema, schema), strings.TrimSpace(id)))
}

func billingPeriodKey(start, end time.Time) string {
	return start.UTC().Format("2006-01-02") + "/" + end.UTC().Format("2006-01-02")
}

func normalizeRetryDelays(delays []time.Duration) []time.Duration {
	out := make([]time.Duration, 0, len(delays))
	for _, delay := range delays {
		if delay > 0 {
			out = append(out, delay)
		}
	}
	if len(out) == 0 {
		return []time.Duration{time.Hour, 6 * time.Hour, 24 * time.Hour}
	}
	if len(out) > 8 {
		return out[:8]
	}
	return out
}

// ClaimDueBillingCycleRetries leases due retry work so multiple scheduler
// replicas cannot advance the same cycle in one poll. RetryBillingCycle sets
// the durable next attempt; the short lease only covers a worker crash.
func (s *Store) ClaimDueBillingCycleRetries(ctx context.Context, now time.Time, limit int) ([]string, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
WITH due AS (
  SELECT id
  FROM %s.billing_cycles
  WHERE status IN ('payment_pending','retry_wait')
    AND next_attempt_at IS NOT NULL
    AND next_attempt_at <= $1
  ORDER BY next_attempt_at
  FOR UPDATE SKIP LOCKED
  LIMIT $2
)
UPDATE %s.billing_cycles c
SET next_attempt_at = $3
FROM due
WHERE c.id = due.id
RETURNING c.id`, schema, schema), now.UTC(), limit, now.UTC().Add(5*time.Minute))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) RetryDueBillingCycles(
	ctx context.Context,
	now time.Time,
	limit int,
	gracePeriod time.Duration,
	retryDelays []time.Duration,
) (int, error) {
	ids, err := s.ClaimDueBillingCycleRetries(ctx, now, limit)
	if err != nil {
		return 0, err
	}
	retried := 0
	for _, id := range ids {
		if _, err := s.RetryBillingCycle(ctx, id, now, gracePeriod, retryDelays); err != nil {
			continue
		}
		retried++
	}
	return retried, nil
}

// ScheduleDueBillingCycles claims due subscriptions with SKIP LOCKED and
// creates one finance/payment action per subscription period. It deliberately
// does not invent a reusable gateway token: the existing provider order is the
// durable hand-off and can be completed by the configured payment flow.
func (s *Store) ScheduleDueBillingCycles(
	ctx context.Context,
	now time.Time,
	limit int,
	gracePeriod time.Duration,
	retryDelays []time.Duration,
) ([]BillingCycle, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}
	if gracePeriod <= 0 {
		gracePeriod = 72 * time.Hour
	}
	retryDelays = normalizeRetryDelays(retryDelays)
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, fmt.Sprintf(`
SELECT %s
FROM %s.tenant_subscriptions s
JOIN %s.package_versions pv ON pv.id = s.package_version_id
WHERE s.status = 'active'
  AND s.next_bill_at IS NOT NULL
  AND s.next_bill_at <= $1
ORDER BY s.next_bill_at
FOR UPDATE OF s SKIP LOCKED
LIMIT $2`, tenantSubscriptionSelect, schema, schema), now, limit)
	if err != nil {
		return nil, err
	}
	subscriptions := make([]TenantSubscription, 0)
	for rows.Next() {
		item, scanErr := scanTenantSubscription(rows)
		if scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		subscriptions = append(subscriptions, *item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	out := make([]BillingCycle, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		periodStart := subscription.CurrentPeriodEnd.UTC()
		periodEnd := addBillingInterval(periodStart, subscription.BillingInterval)
		periodKey := billingPeriodKey(periodStart, periodEnd)
		cycleID := "bcy_" + newStoreID()
		orderID := newPaymentOrderID()
		orderNo := newPaymentOrderNo(subscription.TenantID)
		snapshot, marshalErr := json.Marshal(subscription.CalculationSnapshot)
		if marshalErr != nil {
			return nil, marshalErr
		}

		var insertedID string
		err = tx.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO %s.billing_cycles (
  id, subscription_id, period_key, period_start, period_end, status,
  calculation_snapshot, attempt_count, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,'scheduled',$6::jsonb,0,$7,$7)
ON CONFLICT (subscription_id, period_key) DO NOTHING
RETURNING id`, schema),
			cycleID, subscription.ID, periodKey, periodStart, periodEnd, string(snapshot), actor).Scan(&insertedID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.payment_orders (
  id, tenant_id, package_id, order_no, amount_cents, currency, status,
  provider, payment_method, created_by, updated_by
) VALUES ($1,$2,$3,$4,$5,$6,'pending',$7,$8,$9,$9)`, schema),
			orderID, subscription.TenantID, subscription.PackageID, orderNo,
			subscription.CalculationSnapshot.AmountDueCents, subscription.CalculationSnapshot.Currency,
			subscription.Provider, subscription.PaymentMethod, actor)
		if err != nil {
			return nil, err
		}
		nextAttempt := now.Add(retryDelays[0])
		graceUntil := now.Add(gracePeriod)
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.billing_cycles
SET status = 'payment_pending', order_id = $2, attempt_count = 1,
    next_attempt_at = $3, last_error_code = '', updated_by = $4
WHERE id = $1`, schema), cycleID, orderID, nextAttempt, actor)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_subscriptions
SET status = 'past_due', grace_until = $2, next_bill_at = $3, updated_by = $4
WHERE id = $1 AND status = 'active'`, schema), subscription.ID, graceUntil, periodStart, actor)
		if err != nil {
			return nil, err
		}
		out = append(out, BillingCycle{
			ID: cycleID, SubscriptionID: subscription.ID, TenantID: subscription.TenantID,
			PackageID: subscription.PackageID, PeriodKey: periodKey, PeriodStart: periodStart,
			PeriodEnd: periodEnd, Status: BillingCyclePaymentPending,
			CalculationSnapshot: subscription.CalculationSnapshot, OrderID: orderID,
			AttemptCount: 1, NextAttemptAt: &nextAttempt,
		})
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

// RetryBillingCycle advances the bounded retry state. A paid order is settled;
// a still-pending order is scheduled for the next bounded retry, and exhausted
// cycles enter grace/suspension without duplicating any artifacts.
func (s *Store) RetryBillingCycle(
	ctx context.Context,
	id string,
	now time.Time,
	gracePeriod time.Duration,
	retryDelays []time.Duration,
) (*BillingCycle, error) {
	if s == nil || s.pg == nil {
		return nil, errors.New("postgres unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	if gracePeriod <= 0 {
		gracePeriod = 72 * time.Hour
	}
	retryDelays = normalizeRetryDelays(retryDelays)
	cycle, err := s.GetBillingCycle(ctx, id)
	if err != nil {
		return nil, err
	}
	if cycle.Status == BillingCycleSettled {
		return cycle, nil
	}
	if cycle.OrderID == "" {
		return nil, ErrBillingCycleConflict
	}
	order, err := s.GetPaymentOrderByID(ctx, cycle.OrderID)
	if err != nil {
		return nil, err
	}
	if order.Status == PaymentOrderStatusPaid {
		if err := s.IssuePaymentDocuments(ctx, order.ID); err != nil {
			return nil, err
		}
		if err := s.SettleBillingCycleForPaidOrder(ctx, order.ID); err != nil {
			return nil, err
		}
		return s.GetBillingCycle(ctx, id)
	}

	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var attempt int
	var status string
	var subscriptionID string
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT attempt_count, status, subscription_id
FROM %s.billing_cycles WHERE id = $1 FOR UPDATE`, schema), cycle.ID).
		Scan(&attempt, &status, &subscriptionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBillingCycleNotFound
	}
	if err != nil {
		return nil, err
	}
	if status == BillingCycleSettled {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return s.GetBillingCycle(ctx, cycle.ID)
	}
	attempt++
	if attempt <= len(retryDelays) {
		next := now.Add(retryDelays[attempt-1])
		if order.Status == PaymentOrderStatusFailed {
			_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.payment_orders
SET status = 'pending', transaction_id = '', updated_by = $2
WHERE id = $1 AND status = 'failed'`, schema), order.ID, actor)
			if err != nil {
				return nil, err
			}
		}
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.billing_cycles
SET status = 'retry_wait', attempt_count = $2, next_attempt_at = $3,
    last_error_code = 'PAYMENT_PENDING', updated_by = $4
WHERE id = $1`, schema), cycle.ID, attempt, next, actor)
	} else {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.billing_cycles
SET status = 'failed', attempt_count = $2, next_attempt_at = NULL,
    last_error_code = 'RETRY_EXHAUSTED', updated_by = $3
WHERE id = $1`, schema), cycle.ID, attempt, actor)
		if err == nil {
			var graceUntil *time.Time
			err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT grace_until FROM %s.tenant_subscriptions WHERE id = $1 FOR UPDATE`, schema), subscriptionID).Scan(&graceUntil)
			if err == nil {
				nextStatus := SubscriptionGrace
				if graceUntil == nil {
					value := now.Add(gracePeriod)
					graceUntil = &value
				} else if !graceUntil.After(now) {
					nextStatus = SubscriptionSuspended
				}
				_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_subscriptions
SET status = $2, grace_until = $3, updated_by = $4
WHERE id = $1`, schema), subscriptionID, nextStatus, graceUntil, actor)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetBillingCycle(ctx, cycle.ID)
}

// SettleBillingCycleForPaidOrder advances the subscription exactly once and
// only after both active commercial documents exist.
func (s *Store) SettleBillingCycleForPaidOrder(ctx context.Context, orderID string) error {
	if s == nil || s.pg == nil {
		return errors.New("postgres unavailable")
	}
	order, err := s.GetPaymentOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != PaymentOrderStatusPaid {
		return ErrBillingCycleConflict
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var cycleID string
	err = s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id FROM %s.billing_cycles WHERE order_id = $1`, schema), orderID).Scan(&cycleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	docs, err := s.ListPaymentDocumentsByOrder(ctx, orderID)
	if err != nil {
		return err
	}
	var receiptID, taxInvoiceID string
	for _, doc := range docs {
		if doc.Status != PaymentDocStatusIssued {
			continue
		}
		switch doc.DocType {
		case PaymentDocTypeReceipt:
			receiptID = doc.ID
		case PaymentDocTypeTaxInvoice:
			taxInvoiceID = doc.ID
		}
	}
	if receiptID == "" || taxInvoiceID == "" {
		return ErrBillingCycleConflict
	}

	actor := auditctx.ActorID(ctx)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var subscriptionID, status string
	var periodStart, periodEnd time.Time
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT subscription_id, status, period_start, period_end
FROM %s.billing_cycles WHERE id = $1 FOR UPDATE`, schema), cycleID).
		Scan(&subscriptionID, &status, &periodStart, &periodEnd)
	if err != nil {
		return err
	}
	if status == BillingCycleSettled {
		return tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.billing_cycles
SET status = 'settled', receipt_id = $2, tax_invoice_id = $3,
    next_attempt_at = NULL, last_error_code = '', updated_by = $4
WHERE id = $1`, schema), cycleID, receiptID, taxInvoiceID, actor)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_subscriptions
SET status = 'active', current_period_start = $2, current_period_end = $3,
    next_bill_at = $3, grace_until = NULL, updated_by = $4
WHERE id = $1`, schema), subscriptionID, periodStart, periodEnd, actor)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SettlePaidBillingCycles recovers cycles whose payment callback committed
// before document or scheduler settlement completed.
func (s *Store) SettlePaidBillingCycles(ctx context.Context, limit int) (int, error) {
	if s == nil || s.pg == nil {
		return 0, errors.New("postgres unavailable")
	}
	if limit <= 0 {
		limit = 25
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT c.order_id
FROM %s.billing_cycles c
JOIN %s.payment_orders o ON o.id = c.order_id
WHERE c.status <> 'settled' AND o.status = 'paid'
ORDER BY c.updated_at
LIMIT $1`, schema, schema), limit)
	if err != nil {
		return 0, err
	}
	orderIDs := make([]string, 0)
	for rows.Next() {
		var orderID string
		if err := rows.Scan(&orderID); err != nil {
			rows.Close()
			return 0, err
		}
		orderIDs = append(orderIDs, orderID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()
	settled := 0
	for _, orderID := range orderIDs {
		if err := s.IssuePaymentDocuments(ctx, orderID); err != nil {
			continue
		}
		if err := s.SettleBillingCycleForPaidOrder(ctx, orderID); err == nil {
			settled++
		}
	}
	return settled, nil
}
