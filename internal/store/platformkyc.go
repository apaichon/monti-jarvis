package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

func (s *Store) ApproveTenantKYC(ctx context.Context, tenantID string) (PlatformKYCDecisionResult, error) {
	return s.decideTenantKYC(ctx, tenantID, true, "")
}

func (s *Store) RejectTenantKYC(ctx context.Context, tenantID, reason string) (PlatformKYCDecisionResult, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return PlatformKYCDecisionResult{}, fmt.Errorf("rejection reason is required")
	}
	return s.decideTenantKYC(ctx, tenantID, false, reason)
}

func (s *Store) decideTenantKYC(ctx context.Context, tenantID string, approve bool, reason string) (PlatformKYCDecisionResult, error) {
	if s.pg == nil {
		return PlatformKYCDecisionResult{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	reviewer := auditctx.ActorID(ctx)
	now := time.Now().UTC()

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return PlatformKYCDecisionResult{}, err
	}
	defer tx.Rollback(ctx)

	var tenantStatus string
	err = tx.QueryRow(ctx, fmt.Sprintf(`SELECT status FROM %s.tenants WHERE id = $1 FOR UPDATE`, schema), tenantID).Scan(&tenantStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return PlatformKYCDecisionResult{}, ErrTenantNotFound
		}
		return PlatformKYCDecisionResult{}, err
	}
	if tenantStatus != "pending_kyc" {
		return PlatformKYCDecisionResult{}, ErrKYCReviewConflict
	}

	var kycStatus string
	err = tx.QueryRow(ctx, fmt.Sprintf(`SELECT status FROM %s.tenant_kyc_profiles WHERE tenant_id = $1 FOR UPDATE`, schema), tenantID).Scan(&kycStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return PlatformKYCDecisionResult{}, ErrKYCReviewConflict
		}
		return PlatformKYCDecisionResult{}, err
	}
	if kycStatus != "submitted" {
		return PlatformKYCDecisionResult{}, ErrKYCReviewConflict
	}

	var regStatus, adminEmail, companyName string
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT status, admin_email, company_name FROM %s.tenant_registrations WHERE tenant_id = $1 FOR UPDATE`, schema), tenantID).
		Scan(&regStatus, &adminEmail, &companyName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return PlatformKYCDecisionResult{}, ErrTenantNotFound
		}
		return PlatformKYCDecisionResult{}, err
	}

	out := PlatformKYCDecisionResult{
		TenantID:    tenantID,
		TenantStatus: tenantStatus,
		ReviewedAt:  now,
		ReviewedBy:  reviewer,
		AdminEmail:  adminEmail,
		CompanyName: companyName,
	}

	if approve {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenants SET status = 'active', updated_by = $2, updated_at = now() WHERE id = $1`, schema), tenantID, reviewer)
		if err != nil {
			return PlatformKYCDecisionResult{}, err
		}
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_registrations
SET status = 'approved', rejection_reason = '', reviewed_at = $2, reviewed_by = $3, updated_by = $3, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, now, reviewer)
		if err != nil {
			return PlatformKYCDecisionResult{}, err
		}
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_kyc_profiles
SET status = 'approved', rejection_reason = '', reviewed_at = $2, reviewed_by = $3, updated_by = $3, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, now, reviewer)
		if err != nil {
			return PlatformKYCDecisionResult{}, err
		}
		out.TenantStatus = "active"
		out.RegistrationStatus = "approved"
		out.KYCStatus = "approved"
	} else {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_registrations
SET status = 'rejected', rejection_reason = $2, reviewed_at = $3, reviewed_by = $4, updated_by = $4, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, reason, now, reviewer)
		if err != nil {
			return PlatformKYCDecisionResult{}, err
		}
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_kyc_profiles
SET status = 'rejected', rejection_reason = $2, reviewed_at = $3, reviewed_by = $4, updated_by = $4, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, reason, now, reviewer)
		if err != nil {
			return PlatformKYCDecisionResult{}, err
		}
		out.TenantStatus = "pending_kyc"
		out.RegistrationStatus = "rejected"
		out.KYCStatus = "rejected"
		out.RejectionReason = reason
	}

	if err := tx.Commit(ctx); err != nil {
		return PlatformKYCDecisionResult{}, err
	}
	return out, nil
}